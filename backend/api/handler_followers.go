package api

import (
	"fmt"
	"net/http"

	"brainbook-api/internal/response"
)

// getUserFollowers handles GET /protected/v1/user/{id}/followers
// Returns a list of accepted followers for the specified user.
func (app *Application) getUserFollowers(w http.ResponseWriter, r *http.Request) {
	viewer := contextGetAuthenticatedUser(r)
	pathUserID := r.PathValue("id")

	targetUserID, err := parseStringID(pathUserID)
	if err != nil {
		app.badRequest(w, r, fmt.Errorf("Invalid user ID: %s", pathUserID))
		return
	}

	targetUser, exists, err := app.DB.UserById(targetUserID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	if !exists {
		app.notFound(w, r)
		return
	}

	isSelf := viewer.ID == targetUserID
	if !targetUser.IsPublic && !isSelf {
		isFollower, err := app.DB.IsFollowing(viewer.ID, targetUserID)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
		if !isFollower {
			if err := response.JSON(w, http.StatusOK, map[string]any{"followers": []any{}}); err != nil {
				app.serverError(w, r, err)
				return
			}
			return
		}
	}

	followers, err := app.DB.FollowersByUserID(targetUserID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if err := response.JSON(w, http.StatusOK, map[string]any{"followers": followers}); err != nil {
		app.serverError(w, r, err)
		return
	}
}
