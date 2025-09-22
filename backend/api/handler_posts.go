package api

import (
	"net/http"

	"brainbook-api/internal/response"
)

func (app *Application) fetchPosts(w http.ResponseWriter, r *http.Request) {
	// Retrieve paginated posts from the database
	posts, err := app.DB.GetPosts()
	if err != nil {
		app.serverError(w, r, err)
		return
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
