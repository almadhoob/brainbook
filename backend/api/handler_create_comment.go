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
		Image     []byte              `json:"image"`
		Validator validator.Validator `json:"-"`
	}

	err := request.DecodeJSON(w, r, &input)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	user := contextGetAuthenticatedUser(r)
	if user == nil {
		app.authenticationRequired(w, r)
		return
	}

	input.Validator.Check(validator.NotBlank(input.Content), "Content must not be empty")
	input.Validator.Check(validator.MaxRunes(input.Content, 500), "Content must not exceed 500 characters")

	if input.Validator.HasErrors() {
		app.failedValidation(w, r, input.Validator)
		return
	}

	currentDateTime := time.CurrentTime()

	// Insert the comment into the database
	commentID, err := app.DB.InsertComment(input.PostID, user.ID, input.Content, input.Image, currentDateTime)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	responseData := map[string]any{
		"comment_id": commentID,
		"f_name":     user.FName,
		"l_name":     user.LName,
		"avatar":     user.Avatar,
		"content":    input.Content,
		"image":      input.Image,
		"created_at": currentDateTime,
	}

	err = response.JSON(w, http.StatusCreated, responseData)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
}
