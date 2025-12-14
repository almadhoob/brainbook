package api

import (
	"fmt"
	"net/http"

	"brainbook-api/internal/response"
)

func (app *Application) getPostComments(w http.ResponseWriter, r *http.Request) {
	postIDStr := r.PathValue("post_id")
	if postIDStr == "" {
		app.badRequest(w, r, fmt.Errorf("post ID is required"))
		return
	}

	postID, err := parseStringID(postIDStr)
	if err != nil {
		app.badRequest(w, r, fmt.Errorf("Invalid post ID: %s", postIDStr))
		return
	}

	viewer := contextGetAuthenticatedUser(r)
	canView, err := app.DB.CanUserViewPost(viewer.ID, postID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	if !canView {
		app.Unauthorized(w, r)
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
			"id":             comment.ID,
			"user_id":        comment.UserSummary.ID,
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
