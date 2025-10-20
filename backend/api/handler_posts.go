package api

import (
	"net/http"

	"brainbook-api/internal/response"
)
// for home feed 
func (app *Application) getPosts(w http.ResponseWriter, r *http.Request) {
	contextUser := contextGetAuthenticatedUser(r)
	// Retrieve paginated posts from the database
	posts, err := app.DB.AllPostsByUserID(contextUser.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Can look better, probably
	var postsWithFullName []map[string]any

	for _, post := range posts {
		postsWithFullName = append(postsWithFullName, map[string]any{
			"user_full_name": post.FullName(),
			"user_avatar":    post.Avatar,
			"content":        post.Content,
			"image":          post.File,
			// AI suggets .UTC().Format(time.RFC3339). Not sure what difference it makes.
			"created_at": post.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	// Prepare response with pagination metadata
	responseData := map[string]any{
		"posts": posts,
	}

	// Send the posts as JSON response
	err = response.JSON(w, http.StatusOK, responseData)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
}
