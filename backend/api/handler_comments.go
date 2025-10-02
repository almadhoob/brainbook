package api

import (
	"brainbook-api/internal/response"
	"fmt"
	"net/http"
)

func (app *Application) getPostComments(w http.ResponseWriter, r *http.Request) {
	pathPostID := r.PathValue("id")

	postID, err := parseStringID(pathPostID)
	if err != nil {
		app.badRequest(w, r, fmt.Errorf("Invalid comment ID: %s", pathPostID))
		return
	}

	comments, err := app.DB.CommentsForPost(postID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Can look better, probably
	var commentsWithFullName []map[string]any

	for _, comment := range comments {
		commentsWithFullName = append(commentsWithFullName, map[string]any{
			"user_full_name": comment.FullName(),
			"user_avatar":    comment.Avatar,
			"content":        comment.Content,
			"file":           comment.File,
			// AI suggets .UTC().Format(time.RFC3339). Not sure what difference it makes.
			"created_at": comment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	responseData := map[string]interface{}{
		"comments": commentsWithFullName,
	}

	err = response.JSON(w, http.StatusOK, responseData)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
}
