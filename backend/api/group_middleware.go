package api

import (
	"fmt"
	"net/http"
)

func (app *Application) withGroup(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupIDStr := r.PathValue("group_id")
		if groupIDStr == "" {
			app.badRequest(w, r, fmt.Errorf("group_id parameter is required"))
			return
		}

		groupID, err := parseStringID(groupIDStr)
		if err != nil || groupID <= 0 {
			app.badRequest(w, r, fmt.Errorf("invalid group ID: %s", groupIDStr))
			return
		}

		group, err := app.DB.GroupByID(groupID)
		if err != nil {
			app.notFound(w, r)
			return
		}

		r = contextSetGroup(r, group)
		next(w, r)
	}
}

func (app *Application) requireGroupMember(next http.HandlerFunc) http.HandlerFunc {
	return app.withGroup(func(w http.ResponseWriter, r *http.Request) {
		user := contextGetAuthenticatedUser(r)
		if user == nil {
			app.authenticationRequired(w, r)
			return
		}

		group := contextGetGroup(r)

		// Allow access if user is the group owner
		if group.OwnerID == user.ID {
			next(w, r)
			return
		}

		// Otherwise, check if user is a member
		isMember, err := app.DB.IsGroupMember(group.ID, user.ID)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
		if !isMember {
			app.Unauthorized(w, r)
			return
		}

		next(w, r)
	})
}

func (app *Application) requireGroupOwner(next http.HandlerFunc) http.HandlerFunc {
	return app.withGroup(func(w http.ResponseWriter, r *http.Request) {
		user := contextGetAuthenticatedUser(r)
		if user == nil {
			app.authenticationRequired(w, r)
			return
		}

		group := contextGetGroup(r)
		if group.OwnerID != user.ID {
			app.Unauthorized(w, r)
			return
		}

		next(w, r)
	})
}
