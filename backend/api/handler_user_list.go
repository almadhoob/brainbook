package api

import (
	"net/http"

	"brainbook-api/internal/response"
)

// TO DO: This has to change as the conversation table will do most of the heavy lifting
func (app *Application) getUserList(w http.ResponseWriter, r *http.Request) {
	user := contextGetAuthenticatedUser(r)

	// Get users from database
	users, err := app.DB.UserList(user.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	var usersWithFullName []map[string]any

	for _, listedUser := range users {
		follows, err := app.DB.IsFollowing(user.ID, listedUser.ID)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		followedBy, err := app.DB.IsFollowing(listedUser.ID, user.ID)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		usersWithFullName = append(usersWithFullName, map[string]any{
			"user_id":           listedUser.ID,
			"user_full_name":    listedUser.FullName(),
			"user_avatar":       listedUser.Avatar,
			"last_message_time": listedUser.LastMessageTime,
			"follows":           follows,
			"followed_by":       followedBy,
		})
	}

	responseData := map[string]interface{}{
		"users": usersWithFullName,
	}

	err = response.JSON(w, http.StatusOK, responseData)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
}
