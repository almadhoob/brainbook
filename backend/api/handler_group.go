package api

import (
	"brainbook-api/internal/database"
	"brainbook-api/internal/response"
	"net/http"
)

func (app *Application) getGroups(w http.ResponseWriter, r *http.Request) {

	groups := []database.Group{}
	groups, err := app.DB.AllGroups()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	responseData := map[string]interface{}{
		"groups": groups,
	}

	err = response.JSON(w, http.StatusOK, responseData)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
}


func (app *Application) userGroups(w http.ResponseWriter, r *http.Request) {

	ctx := contextGetAuthenticatedUser(r)
	userID := ctx.ID


	groups := []database.Group{}
	groups, err := app.DB.GroupsByUserID(userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	responseData := map[string]interface{}{
		"groups": groups,
	}

	err = response.JSON(w, http.StatusOK, responseData)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

}


