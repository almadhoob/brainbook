package api

import (
	"net/http"

	"brainbook-api/internal/request"
	"brainbook-api/internal/response"
	time "brainbook-api/internal/time"
	"brainbook-api/internal/validator"
)

func (app *Application) createComment(w http.ResponseWriter, r *http.Request) {
	var input struct {
		PostID    int                 `json:"post_id"`
		Content   string              `json:"content"`
		File      []byte              `json:"file"`
		Validator validator.Validator `json:"-"`
	}

	err := request.DecodeJSON(w, r, &input)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	user := contextGetAuthenticatedUser(r)

	input.Validator.Check(validator.NotBlank(input.Content), "Content must not be empty")
	input.Validator.Check(validator.MaxRunes(input.Content, 500), "Content must not exceed 500 characters")
	if len(input.File) > 0 {
		input.Validator.Check(len(input.File) <= 10_000_000, "File size must be 10MB or less")
	}

	if input.Validator.HasErrors() {
		app.failedValidation(w, r, input.Validator)
		return
	}

	// Ensure commenter can view the target post respecting privacy settings
	canView, err := app.DB.CanUserViewPost(user.ID, input.PostID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	if !canView {
		app.Unauthorized(w, r)
		return
	}

	currentDateTime := time.CurrentTime()

	// Insert the comment into the database
	commentID, err := app.DB.InsertComment(input.PostID, user.ID, input.Content, input.File, currentDateTime)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	responseData := map[string]any{
		"comment_id":     commentID,
		"user_full_name": user.FullName(),
		"user_avatar":    user.Avatar,
		"content":        input.Content,
		"file":           input.File,
		"created_at":     currentDateTime,
	}

	err = response.JSON(w, http.StatusCreated, responseData)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
}
