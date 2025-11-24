package api

import (
	"fmt"
	"net/http"

	"brainbook-api/internal/request"
	"brainbook-api/internal/response"
	t "brainbook-api/internal/time"
	"brainbook-api/internal/validator"
)

func (app *Application) getGroupPostComments(w http.ResponseWriter, r *http.Request) {

	postIDStr := r.PathValue("post_id")
	postID, err := parseStringID(postIDStr)
	if err != nil || postID <= 0 {
		app.badRequest(w, r, fmt.Errorf("invalid post ID: %s", postIDStr))
		return
	}

	comments, err := app.DB.GetCommentsForGroupPost(postID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	responseData := map[string]any{"comments": comments}
	if err := response.JSON(w, http.StatusOK, responseData); err != nil {
		app.serverError(w, r, err)
	}
}

func (app *Application) createGroupPostComment(w http.ResponseWriter, r *http.Request) {
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

	postIDStr := r.PathValue("post_id")
	postID, err := parseStringID(postIDStr)
	if err != nil || postID <= 0 {
		app.badRequest(w, r, fmt.Errorf("invalid post ID: %s", postIDStr))
		return
	}

	isMember, err := app.DB.IsGroupMember(group.ID, userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	if !isMember {
		app.Unauthorized(w, r)
		return
	}

	commentID, err := app.DB.InsertGroupPostComment(input.Content, input.File, t.CurrentTime(), postID, userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	responseData := map[string]any{
		"comment_id": commentID,
	}
	if err := response.JSON(w, http.StatusCreated, responseData); err != nil {
		app.serverError(w, r, err)
	}
}
