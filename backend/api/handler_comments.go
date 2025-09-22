package api

import (
	"brainbook-api/internal/response"
	"fmt"
	"net/http"
)

func (app *Application) fetchPostComments(w http.ResponseWriter, r *http.Request) {
	pathPostID := r.PathValue("id")

	postID, err := app.parseStringID(pathPostID)
	if err != nil {
		app.badRequest(w, r, fmt.Errorf("Invalid user ID: %s", pathPostID))
		return
	}

	comments, err := app.DB.GetCommentsForPost(postID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	responseData := map[string]any{
		"comments": comments,
	}

	err = response.JSON(w, http.StatusOK, responseData)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
}
