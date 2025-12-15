package api

import (
	"net/http"

	"brainbook-api/internal/response"
)

// getCurrentUser returns minimal info about the authenticated user
func (app *Application) getCurrentUser(w http.ResponseWriter, r *http.Request) {
	user := contextGetAuthenticatedUser(r)
	if user == nil {
		app.authenticationRequired(w, r)
		return
	}

	payload := map[string]any{
		"user_id":  user.ID,
		"email":    user.Email,
		"f_name":   user.FName,
		"l_name":   user.LName,
		"nickname": user.Nickname,
	}

	if err := response.JSON(w, http.StatusOK, payload); err != nil {
		app.serverError(w, r, err)
	}
}
