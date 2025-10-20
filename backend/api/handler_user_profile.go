package api

import (
	"fmt"
	"net/http"

	"brainbook-api/internal/database"
	"brainbook-api/internal/response"
)

// fetchUserProfile handles GET /protected/v1/user/me
// Returns current user's profile information
func (app *Application) getUserProfile(w http.ResponseWriter, r *http.Request) {
	// Retrieves user from context
	contextUser := contextGetAuthenticatedUser(r)
	pathUserID := r.PathValue("id")

	targetUserID, err := parseStringID(pathUserID)
	if err != nil {
		app.badRequest(w, r, fmt.Errorf("Invalid user ID: %s", pathUserID))
		return
	}

	isUserIDMatching := contextUser.IsUserIDMatching(targetUserID)

	var profileUser *database.User
	var posts []database.Post
	var pendingFollowRequestsCount int

	if isUserIDMatching {
		// Target user exists
		targetUser, exists, err := app.DB.UserById(targetUserID)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
		if !exists {
			app.notFound(w, r)
			return
		}

		profileUser = targetUser

	} else {
		// Viewer is the same as target user
		profileUser = contextUser
		targetUserID = contextUser.ID

	pendingFollowRequestsCount, err = app.DB.PendingFollowRequestsCount(targetUserID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	}
// Retrieve posts visible to the context user from the target user including self
	posts, err = app.DB.PostsVisibleFromUser(contextUser.ID, targetUserID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	/* If profile is private, check if context user if a follower (accepted in follow_request table).
	   If not, respond with 401 Unauthorized. Refer to:
	   github.com/0xdod/go-realworld/blob/master/conduit/user.go#L39*/

	// TO DO: Create the two functions below. Kindly do not use AI as it is not needed!
	followerCount, err := app.DB.FollowerCountByUserID(targetUserID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	followingCount, err := app.DB.FollowingCountByUserID(targetUserID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	

	userProfileResponse := map[string]any{
		"user_id":                       profileUser.ID,
		"full_name":                     profileUser.FullName(),
		"email":                         profileUser.Email,
		"dob":                           profileUser.DOB,
		"follower_count":                followerCount,
		"following_count":               followingCount,
		"posts":                         posts,
		"pending_follow_requests_count": pendingFollowRequestsCount,
		/*Used in the frontend to determine when
		  to display the follow/unfollow button*/
		"is_self": isUserIDMatching,
	}

	// Potentially empty values
	if profileUser.Avatar != nil {
		userProfileResponse["avatar"] = profileUser.Avatar
	}

	if profileUser.Nickname != "" {
		userProfileResponse["nickname"] = profileUser.Nickname
	}

	if profileUser.Bio != "" {
		userProfileResponse["bio"] = profileUser.Bio
	}

	err = response.JSON(w, http.StatusOK, userProfileResponse)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
}
