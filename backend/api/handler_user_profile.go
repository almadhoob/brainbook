package api

import (
	"fmt"
	"net/http"

	"brainbook-api/internal/response"
)

// fetchUserProfile handles GET /protected/v1/user/me
// Returns current user's profile information
func (app *Application) getUserProfile(w http.ResponseWriter, r *http.Request) {
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

	// Enforce private profile visibility
	if !targetUser.IsPublic && !isSelf {
		isFollower, err := app.DB.IsFollowing(viewer.ID, targetUserID)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
		if !isFollower {
			app.Unauthorized(w, r)
			return
		}
	}

	posts, err := app.DB.PostsVisibleFromUser(viewer.ID, targetUserID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	followers, err := app.DB.FollowersByUserID(targetUserID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	following, err := app.DB.FollowingByUserID(targetUserID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	followRequestStatus := ""
	if !isSelf {
		status, exists, err := app.DB.FollowRequestStatus(viewer.ID, targetUserID)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
		if exists {
			followRequestStatus = status
		}
	}

	pendingFollowRequestsCount := 0
	if isSelf {
		pendingFollowRequestsCount, err = app.DB.PendingFollowRequestsCount(targetUserID)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
	}

	userProfileResponse := map[string]any{
		"user_id":                       targetUser.ID,
		"full_name":                     targetUser.FullName(),
		"email":                         targetUser.Email,
		"dob":                           targetUser.DOB,
		"is_public":                     targetUser.IsPublic,
		"followers":                     followers,
		"following":                     following,
		"posts":                         posts,
		"pending_follow_requests_count": pendingFollowRequestsCount,
		"is_self":                       isSelf,
		"follow_request_status":         followRequestStatus,
	}

	if targetUser.Avatar != nil {
		userProfileResponse["avatar"] = targetUser.Avatar
	}
	if targetUser.Nickname != "" {
		userProfileResponse["nickname"] = targetUser.Nickname
	}
	if targetUser.Bio != "" {
		userProfileResponse["bio"] = targetUser.Bio
	}

	if err := response.JSON(w, http.StatusOK, userProfileResponse); err != nil {
		app.serverError(w, r, err)
		return
	}
}
