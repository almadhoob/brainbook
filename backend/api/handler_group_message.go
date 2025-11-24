package api

import (
	"net/http"

	"brainbook-api/internal/response"
)

func (app *Application) getGroupMessages(w http.ResponseWriter, r *http.Request) {
	group := contextGetGroup(r)
	limit := parseQueryInt(r, "limit", 50)
	offset := parseQueryInt(r, "offset", 0)

	messages, err := app.DB.GroupMessages(group.ID, limit, offset)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if err := response.JSON(w, http.StatusOK, map[string]any{"messages": messages}); err != nil {
		app.serverError(w, r, err)
	}
}
