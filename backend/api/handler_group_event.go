package api

import (
	"fmt"
	"net/http"

	"brainbook-api/internal/request"
	"brainbook-api/internal/response"
)

func (app *Application) createGroupEvent(w http.ResponseWriter, r *http.Request) {
	group := contextGetGroup(r)

	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Time        string `json:"time"`
	}

	if err := request.DecodeJSON(w, r, &input); err != nil {
		app.badRequest(w, r, err)
		return
	}

	if input.Title == "" || input.Time == "" {
		app.badRequest(w, r, fmt.Errorf("title and time are required"))
		return
	}

	ctxUser := contextGetAuthenticatedUser(r)

	eventID, err := app.DB.InsertGroupEvent(ctxUser.ID, input.Title, input.Description, input.Time, group.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// notify members
	members, err := app.DB.GroupMembersByGroupID(group.ID)
	if err == nil {
		for _, member := range members {
			if member.ID == ctxUser.ID {
				continue
			}
			app.notifyUser(member.ID, NotificationTypeGroupEvent, map[string]interface{}{
				"group_id": group.ID,
				"event_id": eventID,
				"title":    input.Title,
			})
		}
	}

	if err := response.JSON(w, http.StatusCreated, map[string]any{"event_id": eventID}); err != nil {
		app.serverError(w, r, err)
	}
}

func (app *Application) listGroupEvents(w http.ResponseWriter, r *http.Request) {
	group := contextGetGroup(r)

	events, err := app.DB.GetGroupEvents(group.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if err := response.JSON(w, http.StatusOK, map[string]any{"events": events}); err != nil {
		app.serverError(w, r, err)
	}
}

func (app *Application) rsvpGroupEvent(w http.ResponseWriter, r *http.Request) {
	user := contextGetAuthenticatedUser(r)
	group := contextGetGroup(r)

	eventIDStr := r.PathValue("event_id")
	eventID, err := parseStringID(eventIDStr)
	if err != nil || eventID <= 0 {
		app.badRequest(w, r, fmt.Errorf("invalid event id: %s", eventIDStr))
		return
	}

	event, exists, err := app.DB.EventByID(eventID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	if !exists || event.GroupID != group.ID {
		app.notFound(w, r)
		return
	}

	var input struct {
		Response string `json:"response"`
	}
	if err := request.DecodeJSON(w, r, &input); err != nil {
		app.badRequest(w, r, err)
		return
	}

	var going bool
	switch input.Response {
	case "going":
		going = true
	case "not_going":
		going = false
	default:
		app.badRequest(w, r, fmt.Errorf("response must be 'going' or 'not_going'"))
		return
	}

	if err := app.DB.UpsertEventRSVP(eventID, user.ID, going); err != nil {
		app.serverError(w, r, err)
		return
	}

	status := "not_going"
	if going {
		status = "going"
	}

	_ = response.JSON(w, http.StatusOK, map[string]any{"status": status})
}
