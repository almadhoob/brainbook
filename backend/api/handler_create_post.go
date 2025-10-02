package api

import (
	"net/http"

	"brainbook-api/internal/request"
	"brainbook-api/internal/response"
	t "brainbook-api/internal/time"
	"brainbook-api/internal/validator"
)

func (app *Application) createPost(w http.ResponseWriter, r *http.Request) {
	// Define the input structure to decode JSON
	var input struct {
		Content   string              `json:"content"`
		File      []byte              `json:"file"`
		Validator validator.Validator `json:"-"`
	}

	// Decode the JSON request body
	err := request.DecodeJSON(w, r, &input)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	// Get the authenticated user from context
	user := contextGetAuthenticatedUser(r)

	input.Validator.CheckField(validator.NotBlank(input.Content), "post-content", "Content must not be empty")
	input.Validator.CheckField(validator.MaxRunes(input.Content, 500), "post-content", "Content must not exceed 500 characters")

	if input.Validator.HasErrors() {
		app.failedValidation(w, r, input.Validator)
		return
	}

	currentDateTime := t.CurrentTime()

	// Insert the post into the database
	postID, err := app.DB.InsertPost(user.ID, input.Content, input.File, currentDateTime)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Respond with the created post
	responseData := map[string]any{
		"post_id":    postID,
		"user_full_name":     user.FullName(),
		"user_avatar":     user.Avatar,
		"content":    input.Content,
		"file:":      input.File,
		"created_at": currentDateTime,
	}

	err = response.JSON(w, http.StatusCreated, responseData)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
}
