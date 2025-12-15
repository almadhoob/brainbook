package api

import (
	"fmt"
	"net/http"

	"brainbook-api/internal/request"
	"brainbook-api/internal/response"
	time "brainbook-api/internal/time"
	"brainbook-api/internal/validator"
)

func (app *Application) createComment(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Content   string              `json:"content"`
		File      []byte              `json:"file"`
		Validator validator.Validator `json:"-"`
	}

	err := request.DecodeJSON(w, r, &input)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	postIDStr := r.PathValue("post_id")
	postID, err := parseStringID(postIDStr)
	if err != nil || postID <= 0 {
		app.badRequest(w, r, fmt.Errorf("invalid post ID: %s", postIDStr))
		return
	}

	user := contextGetAuthenticatedUser(r)

	input.Validator.CheckField(validator.NotBlank(input.Content), "content", "Content must not be empty")
	input.Validator.CheckField(validator.MaxRunes(input.Content, 500), "content", "Content must not exceed 500 characters")
	if len(input.File) > 0 {
		input.Validator.CheckField(len(input.File) <= 10_000_000, "file", "File size must be 10MB or less")
	}

	if input.Validator.HasErrors() {
		app.failedValidation(w, r, input.Validator)
		return
	}

	// Ensure commenter can view the target post respecting privacy settings
	canView, err := app.DB.CanUserViewPost(user.ID, postID)
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
	commentID, err := app.DB.InsertComment(postID, user.ID, input.Content, input.File, currentDateTime)
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
