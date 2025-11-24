package api

import (
	"fmt"
	"net/http"

	"brainbook-api/internal/request"
	"brainbook-api/internal/response"
	"brainbook-api/internal/validator"
)

func (app *Application) getMembers(w http.ResponseWriter, r *http.Request) {
	ctx := contextGetAuthenticatedUser(r)
	userID := ctx.ID
	group := contextGetGroup(r)

	ismember, err := app.DB.IsGroupMember(group.ID, userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if !ismember {
		app.Unauthorized(w, r)
		return
	}

	members, err := app.DB.GroupMembersByGroupID(group.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	responseData := map[string]interface{}{
		"members": members,
	}

	err = response.JSON(w, http.StatusOK, responseData)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
}

// used by user to request to join a group
func (app *Application) joinGroupRequest(w http.ResponseWriter, r *http.Request) {

	ctx := contextGetAuthenticatedUser(r)
	userID := ctx.ID
	group := contextGetGroup(r)

	ismember, err := app.DB.IsGroupMember(userID, group.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if ismember {
		_ = response.JSON(w, http.StatusOK, map[string]any{"status": "member"})
		return
	}

	if group.OwnerID == userID {
		_ = response.JSON(w, http.StatusOK, map[string]any{"status": "owner"})
		return
	}

	exists, pending, err := app.DB.RequestExistsAndPending(group.ID, userID, group.OwnerID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// If request already pending, return status JSON
	if exists && pending {
		_ = response.JSON(w, http.StatusOK, map[string]any{"status": "pending"})
		return
	}

	existsOpp, pendingOpp, err := app.DB.RequestExistsAndPending(group.ID, group.OwnerID, userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	if existsOpp {
		status := "pending"
		if !pendingOpp {
			status = "processed"
		}
		_ = response.JSON(w, http.StatusOK, map[string]any{"status": status})
		return
	}

	// Create a new request and return pending status
	err = app.DB.InsertJoinRequest(group.ID, userID, group.OwnerID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.notifyUser(group.OwnerID, NotificationTypeGroupJoin, map[string]interface{}{
		"group_id":     group.ID,
		"requester_id": userID,
	})

	_ = response.JSON(w, http.StatusCreated, map[string]any{"status": "pending"})
}

// used by group owner to invite users to join group
func (app *Application) SendGroupInvite(w http.ResponseWriter, r *http.Request) {

	var input struct {
		TargetUserID int                 `json:"target_user_id"`
		Validator    validator.Validator `json:"-"`
	}
	err := request.DecodeJSON(w, r, &input)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	input.Validator.CheckField(input.TargetUserID > 0, "target_user_id", "Target user is required")
	if input.Validator.HasErrors() {
		app.failedValidation(w, r, input.Validator)
		return
	}

	//requester
	ctx := contextGetAuthenticatedUser(r)
	userID := ctx.ID
	group := contextGetGroup(r)
	if input.TargetUserID == userID {
		app.badRequest(w, r, fmt.Errorf("cannot invite yourself"))
		return
	}

	_, found, err := app.DB.UserById(input.TargetUserID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	if !found {
		app.notFound(w, r)
		return
	}

	isTargetMember, err := app.DB.IsGroupMember(group.ID, input.TargetUserID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	if isTargetMember {
		_ = response.JSON(w, http.StatusOK, map[string]any{"status": "member"})
		return
	}
	if group.OwnerID != userID {
		app.Unauthorized(w, r)
		return
	}

	exists, pending, err := app.DB.RequestExistsAndPending(group.ID, userID, input.TargetUserID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	if exists {
		status := "pending"
		if !pending {
			status = "processed"
		}
		_ = response.JSON(w, http.StatusOK, map[string]any{"status": status})
		return
	}

	err = app.DB.InsertJoinRequest(group.ID, userID, input.TargetUserID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.notifyUser(input.TargetUserID, NotificationTypeGroupInvite, map[string]interface{}{
		"group_id":   group.ID,
		"inviter_id": userID,
	})

	_ = response.JSON(w, http.StatusCreated, map[string]any{"status": "invite sent"})
}
