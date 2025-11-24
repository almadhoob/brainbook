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
