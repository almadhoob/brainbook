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

// Get groups that the authenticated user is a member of
func (app *Application) userGroups(w http.ResponseWriter, r *http.Request) {

	ctx := contextGetAuthenticatedUser(r)
	userID := ctx.ID

	groups := []database.Group{}
	groups, err := app.DB.GroupsByUserID(userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	//Should validate if user has no groups
	responseData := map[string]interface{}{
		"groups": groups,
	}

	err = response.JSON(w, http.StatusOK, responseData)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

}


func (app *Application) groupDetails(w http.ResponseWriter, r *http.Request) {
	groupIDStr := r.PathValue("group_id")
	groupID, err := parseStringID(groupIDStr)
	if err != nil || groupID <= 0 {
		app.badRequest(w, r, err)
		return
	}

	group,  err := app.DB.GroupByID(groupID)
	if err != nil {
		app.notFound(w, r)
		return
	}

	responseData := map[string]interface{}{
		"group": group,
	}

	err = response.JSON(w, http.StatusOK, responseData)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
}

