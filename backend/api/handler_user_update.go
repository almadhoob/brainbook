package api

import (
	"net/http"

	"brainbook-api/internal/request"
	"brainbook-api/internal/validator"
)

func (app *Application) updateProfile(w http.ResponseWriter, r *http.Request) {
	contextUser := contextGetAuthenticatedUser(r)
	targetUserID := contextUser.ID

	// Use pointers to detect presence vs. absence
	var input struct {
		Nickname  *string             `json:"nickname,omitempty"`
		Bio       *string             `json:"bio,omitempty"`
		Avatar    *[]byte             `json:"avatar,omitempty"` // base64 in JSON -> []byte
		IsPublic  *bool               `json:"is_public,omitempty"`
		Validator validator.Validator `json:"-"`
	}

	if err := request.DecodeJSON(w, r, &input); err != nil {
		app.badRequest(w, r, err)
		return
	}

	// Validate only provided, non-empty values
	v := input.Validator

	if input.Nickname != nil {
		v.CheckField(validator.MaxRunes(*input.Nickname, 50), "nickname", "Nickname must be 50 characters or less")
		// Do not update if empty or whitespace
		if !validator.NotBlank(*input.Nickname) {
			input.Nickname = nil
		}
	}
	if input.Bio != nil {
		v.CheckField(validator.MaxRunes(*input.Bio, 500), "bio", "Bio limit exceeded (500 characters)")
		// Do not update if empty or whitespace
		if !validator.NotBlank(*input.Bio) {
			input.Bio = nil
		}
	}
	if input.Avatar != nil {
		const maxAvatar = 5_000_000 // 5 MB
		if len(*input.Avatar) == 0 {
			// Do not update if empty payload
			input.Avatar = nil
		} else {
			v.CheckField(len(*input.Avatar) <= maxAvatar, "avatar", "Avatar size limit exceeded (5MB)")
		}
	}

	if v.HasErrors() {
		app.failedValidation(w, r, v)
		return
	}

	// If nothing to update, return 204 without hitting DB
	if input.Nickname == nil && input.Bio == nil && input.Avatar == nil && input.IsPublic == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if input.Nickname != nil {
		if err := app.DB.UpdateNickname(targetUserID, *input.Nickname); err != nil {
			app.serverError(w, r, err)
			return
		}
	}
	if input.Bio != nil {
		if err := app.DB.UpdateBio(targetUserID, *input.Bio); err != nil {
			app.serverError(w, r, err)
			return
		}
	}
	if input.Avatar != nil {
		if err := app.DB.UpdateAvatar(targetUserID, *input.Avatar); err != nil {
			app.serverError(w, r, err)
			return
		}
	}
	if input.IsPublic != nil {
		if err := app.DB.UpdatePrivacy(targetUserID, *input.IsPublic); err != nil {
			app.serverError(w, r, err)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}
