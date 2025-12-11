package api

import (
	"net/http"

	"brainbook-api/internal/response"
)

// getSessionProfile returns the authenticated user's lightweight profile data.
func (app *Application) getSessionProfile(w http.ResponseWriter, r *http.Request) {
	user := contextGetAuthenticatedUser(r)
	if user == nil {
		app.authenticationRequired(w, r)
		return
	}

	payload := map[string]any{
		"user_id":   user.ID,
		"full_name": user.FullName(),
		"email":     user.Email,
	}

	if len(user.Avatar) > 0 {
		payload["avatar"] = user.Avatar
	}

	if err := response.JSON(w, http.StatusOK, payload); err != nil {
		app.serverError(w, r, err)
		return
	}
}
