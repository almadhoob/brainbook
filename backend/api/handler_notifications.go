package api

import (
	"fmt"
	"net/http"

	"brainbook-api/internal/response"
)

func (app *Application) getNotifications(w http.ResponseWriter, r *http.Request) {
	user := contextGetAuthenticatedUser(r)
	if user == nil {
		app.authenticationRequired(w, r)
		return
	}

	includeRead := r.URL.Query().Get("all") == "1"
	notifications, err := app.DB.NotificationsByUser(user.ID, includeRead)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if err := response.JSON(w, http.StatusOK, map[string]any{"notifications": notifications}); err != nil {
		app.serverError(w, r, err)
	}
}

func (app *Application) markNotificationRead(w http.ResponseWriter, r *http.Request) {
	user := contextGetAuthenticatedUser(r)
	if user == nil {
		app.authenticationRequired(w, r)
		return
	}

	notifIDStr := r.PathValue("notification_id")
	notifID, err := parseStringID(notifIDStr)
	if err != nil || notifID <= 0 {
		app.badRequest(w, r, fmt.Errorf("invalid notification id: %s", notifIDStr))
		return
	}

	if err := app.DB.MarkNotificationRead(notifID, user.ID); err != nil {
		app.serverError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
