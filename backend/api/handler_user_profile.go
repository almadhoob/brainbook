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

	if isUserIDMatching {
		// Checks if the target user exists.
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
		profileUser = contextUser
	}

	/* If profile is private, check if context user if a follower (accepted in follow_request table).
	   If not, respond with 401 Unauthorized. Refer to:
	   github.com/0xdod/go-realworld/blob/master/conduit/user.go#L39*/

	// TO DO: Create the two functions below. Kindly do not use AI as it is not needed!
	followerCount, err := app.DB.UserFollowerCount(targetUserID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	followingCount, err := app.DB.UserFollowingCount(targetUserID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	posts, err := app.DB.PostsByUserID(targetUserID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	userProfileResponse := map[string]any{
		"user_id":         profileUser.ID,
		"full_name":       profileUser.FullName(),
		"email":           profileUser.Email,
		"dob":             profileUser.DOB,
		"follower_count":  followerCount,
		"following_count": followingCount,
		"posts":           posts,
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
