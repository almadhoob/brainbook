package api

import (
	"brainbook-api/internal/request"
	"brainbook-api/internal/response"
	"brainbook-api/internal/validator"
	"net/http"
)

func (app *Application) getMembers(w http.ResponseWriter, r *http.Request) {
	ctx := contextGetAuthenticatedUser(r)
	userID := ctx.ID

	groupIDStr := r.PathValue("group_id")
	groupID, err := parseStringID(groupIDStr)
	if err != nil || groupID <= 0 {
		app.badRequest(w, r, err)
		return
	}

	// Check if group exists
	_, err = app.DB.GroupByID(groupID)
	if err != nil {
		app.notFound(w, r)
		return
	}

	ismember, err := app.DB.IsGroupMember(userID, groupID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if !ismember {
		app.Unauthorized(w, r)
		return
	}

	members, err := app.DB.GroupMembersByGroupID(groupID)
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

	groupIDStr := r.PathValue("group_id")
	groupID, err := parseStringID(groupIDStr)
	if err != nil || groupID <= 0 {
		app.badRequest(w, r, err)
		return
	}

	// Check if group exists
	group, err := app.DB.GroupByID(groupID)
	if err != nil {
		app.notFound(w, r)
		return
	}

	ismember, err := app.DB.IsGroupMember(userID, groupID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if ismember {
		_ = response.JSON(w, http.StatusOK, map[string]any{"status": "member"})
		return
	}

	exists, pending, err := app.DB.RequestExistsAndPending(groupID, userID, group.OwnerID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// If request already pending, return status JSON
	if exists && pending {
		_ = response.JSON(w, http.StatusOK, map[string]any{"status": "pending"})
		return
	}

	// Create a new request and return pending status
	err = app.DB.InsertJoinRequest(groupID, userID, group.OwnerID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

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
	//requester
	ctx := contextGetAuthenticatedUser(r)
	userID := ctx.ID

	groupIDStr := r.PathValue("group_id")
	groupID, err := parseStringID(groupIDStr)
	if err != nil || groupID <= 0 {
		app.badRequest(w, r, err)
		return
	}

	// Check if group exists
	group, err := app.DB.GroupByID(groupID)
	if err != nil {
		app.notFound(w, r)
		return
	}
	_, err = app.DB.IsGroupMember(groupID, input.TargetUserID)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}
	if group.OwnerID != userID {
		app.Unauthorized(w, r)
		return
	}

	err = app.DB.InsertJoinRequest(groupID, userID, input.TargetUserID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	_ = response.JSON(w, http.StatusCreated, map[string]any{"status": "invite sent"})
}
