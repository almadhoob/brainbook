package api

import (
	"brainbook-api/internal/request"
	"brainbook-api/internal/response"
	t "brainbook-api/internal/time"
	"brainbook-api/internal/validator"
	"net/http"
)

func (app *Application) groupPosts(w http.ResponseWriter, r *http.Request) {
	group := contextGetGroup(r)

	posts, err := app.DB.GetGroupPosts(group.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	responseData := map[string]interface{}{
		"posts": posts,
	}

	err = response.JSON(w, http.StatusOK, responseData)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
}

func (app *Application) groupPostCreate(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Content   string              `json:"content"`
		File      []byte              `json:"file"`
		Validator validator.Validator `json:"-"`
	}

	if err := request.DecodeJSON(w, r, &input); err != nil {
		app.badRequest(w, r, err)
		return
	}

	input.Validator.CheckField(validator.NotBlank(input.Content), "content", "Content must not be empty")
	input.Validator.CheckField(validator.MaxRunes(input.Content, 500), "content", "Content must not exceed 500 characters")
	if input.Validator.HasErrors() {
		app.failedValidation(w, r, input.Validator)
		return
	}

	ctx := contextGetAuthenticatedUser(r)
	userID := ctx.ID

	group := contextGetGroup(r)

	//content string, image []byte, currentDateTime string, userID int, groupID int
	postID, err := app.DB.InsertGroupPost(input.Content, input.File, t.CurrentTime(), userID, group.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	responseData := map[string]interface{}{
		"post_id": postID,
	}

	err = response.JSON(w, http.StatusCreated, responseData)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
}
