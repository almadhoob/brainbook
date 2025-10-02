package api

import (
	"net/http"

	"brainbook-api/internal/response"
)

// TO DO: This has to change as the conversation table will do most of the heavy lifting
type UserListResponse struct {
	ID              int     `json:"id"`
	FullName        string  `json:"full_name"`
	Status          int     `json:"status"` // 0 = offline, 1 = online
	LastMessageTime *string `json:"last_message_time"`
}

func (app *Application) getUserList(w http.ResponseWriter, r *http.Request) {
	user := contextGetAuthenticatedUser(r)

	// Get users from database
	users, err := app.DB.UserList(user.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Convert to response format (status will be updated by frontend from WebSocket cache)
	var userList []UserListResponse
	for _, User := range users {
		userList = append(userList, UserListResponse{
			ID:              User.ID,
			FullName:        User.FName + " " + User.LName,
			Status:          0, // Default to offline, frontend will update from WebSocket cache
			LastMessageTime: User.LastMessageTime,
		})
	}

	// Get total count excluding current user (consistent with pagination query)
	totalCount, err := app.DB.TotalUserCountExcludingUser(user.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Get online count from WebSocket manager (excluding current user)
	onlineCount := 0
	offlineCount := totalCount
	if app.WSManager != nil {
		onlineUserIDs := app.WSManager.GetOnlineUserIDs()

		// Exclude current user from online count
		for _, userID := range onlineUserIDs {
			if userID != user.ID {
				onlineCount++
			}
		}

		offlineCount = totalCount - onlineCount

		// Ensure counts are not negative
		if onlineCount < 0 {
			onlineCount = 0
		}
		if offlineCount < 0 {
			offlineCount = 0
		}
	}

	// Prepare response with pagination metadata (same pattern as posts)
	responseData := map[string]interface{}{
		"users":         userList,
		"online_count":  onlineCount,
		"offline_count": offlineCount,
	}

	// Return the paginated user list as JSON
	err = response.JSON(w, http.StatusOK, responseData)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
}
