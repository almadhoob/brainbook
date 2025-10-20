package api

import (
	"brainbook-api/internal/request"
	"brainbook-api/internal/validator"
	"net/http"
)

func (app *Application) createGroup(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Title               string `json:"title"`
		Description         string `json:"description"`
		validator.Validator `json:"-"`
	}
	ctx := contextGetAuthenticatedUser(r)
	userID := ctx.ID

	if err := request.DecodeJSON(w, r, &input); err != nil {
		app.badRequest(w, r, err)
		return
	}

	input.Check(validator.NotBlank(input.Title), "Title must not be empty")
	input.Check(validator.MaxRunes(input.Title, 100), "Title must not exceed 100 characters")
	input.Check(validator.NotBlank(input.Description), "Description must not be empty")

	input.Check(validator.MaxRunes(input.Description, 500), "Description must not exceed 500 characters")

	if input.HasErrors() {
		app.failedValidation(w, r, input.Validator)
		return
	}
	if _, err := app.DB.InsertGroup(userID, input.Title, input.Description); err != nil {
		app.serverError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
