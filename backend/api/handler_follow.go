package api

import (
	"fmt"
	"net/http"

	"brainbook-api/internal/request"
	"brainbook-api/internal/response"
)

func (app *Application) sendFollowRequest(w http.ResponseWriter, r *http.Request) {
	user := contextGetAuthenticatedUser(r)
	targetIDStr := r.PathValue("user_id")
	targetID, err := parseStringID(targetIDStr)
	if err != nil || targetID <= 0 {
		app.badRequest(w, r, fmt.Errorf("invalid user id: %s", targetIDStr))
		return
	}
	if targetID == user.ID {
		app.badRequest(w, r, fmt.Errorf("cannot follow yourself"))
		return
	}

	targetUser, found, err := app.DB.UserById(targetID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	if !found {
		app.notFound(w, r)
		return
	}

	if existing, exists, err := app.DB.FollowRequestBetween(user.ID, targetID); err != nil {
		app.serverError(w, r, err)
		return
	} else if exists {
		_ = response.JSON(w, http.StatusOK, map[string]any{"status": existing.Status})
		return
	}

	status := "pending"
	if targetUser.IsPublic {
		status = "accepted"
	}

	req, err := app.DB.CreateFollowRequest(user.ID, targetID, status)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if status == "pending" {
		app.notifyUser(targetID, NotificationTypeFollowRequest, map[string]interface{}{
			"request_id":   req.ID,
			"requester_id": user.ID,
		})
	} else {
		app.notifyUser(targetID, NotificationTypeFollowRequest, map[string]interface{}{
			"request_id": req.ID,
			"status":     status,
		})
	}

	_ = response.JSON(w, http.StatusCreated, map[string]any{"status": status})
}

func (app *Application) respondFollowRequest(w http.ResponseWriter, r *http.Request) {
	user := contextGetAuthenticatedUser(r)

	requestIDStr := r.PathValue("request_id")
	requestID, err := parseStringID(requestIDStr)
	if err != nil || requestID <= 0 {
		app.badRequest(w, r, fmt.Errorf("invalid request id: %s", requestIDStr))
		return
	}

	var input struct {
		Action string `json:"action"`
	}
	if err := request.DecodeJSON(w, r, &input); err != nil {
		app.badRequest(w, r, err)
		return
	}

	var newStatus string
	switch input.Action {
	case "accept":
		newStatus = "accepted"
	case "decline":
		newStatus = "declined"
	default:
		app.badRequest(w, r, fmt.Errorf("invalid action"))
		return
	}

	fr, err := app.DB.FollowRequestByID(requestID)
	if err != nil {
		app.notFound(w, r)
		return
	}
	if fr.TargetID != user.ID {
		app.Unauthorized(w, r)
		return
	}

	if err := app.DB.UpdateFollowRequestStatus(fr.ID, newStatus); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.notifyUser(fr.RequesterID, NotificationTypeFollowRequest, map[string]interface{}{
		"request_id": fr.ID,
		"status":     newStatus,
	})

	_ = response.JSON(w, http.StatusOK, map[string]any{"status": newStatus})
}

// getPendingFollowRequests returns pending follow requests for the authenticated user.
func (app *Application) getPendingFollowRequests(w http.ResponseWriter, r *http.Request) {
	user := contextGetAuthenticatedUser(r)

	reqs, err := app.DB.PendingFollowRequests(user.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	_ = response.JSON(w, http.StatusOK, map[string]any{"requests": reqs})
}

// unfollowUser removes an accepted follow relation.
func (app *Application) unfollowUser(w http.ResponseWriter, r *http.Request) {
	user := contextGetAuthenticatedUser(r)

	targetIDStr := r.PathValue("user_id")
	targetID, err := parseStringID(targetIDStr)
	if err != nil || targetID <= 0 {
		app.badRequest(w, r, fmt.Errorf("invalid user id: %s", targetIDStr))
		return
	}
	if targetID == user.ID {
		app.badRequest(w, r, fmt.Errorf("cannot unfollow yourself"))
		return
	}

	if err := app.DB.DeleteFollow(user.ID, targetID); err != nil {
		app.serverError(w, r, err)
		return
	}

	_ = response.JSON(w, http.StatusOK, map[string]any{"status": "unfollowed"})
}
