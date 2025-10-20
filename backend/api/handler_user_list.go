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

	for _, user := range users {
		usersWithFullName = append(usersWithFullName, map[string]any{
			"user_full_name": user.FullName(),
			"user_avatar":    user.Avatar,
			"last_message_time": user.LastMessageTime,
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
