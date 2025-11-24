package api

import (
	"context"
	"net/http"

	"brainbook-api/internal/database"
)

// declaring custom contextKey type
type contextKey string

// creating constant with the type contextKey
const (
	authenticatedUserContextKey = contextKey("authenticatedUser")
	groupContextKey            = contextKey("group")
)

// sets the value in request context, using the constant above
func contextSetAuthenticatedUser(r *http.Request, user *database.User) *http.Request {
	ctx := context.WithValue(r.Context(), authenticatedUserContextKey, user)
	return r.WithContext(ctx)
}

func contextGetAuthenticatedUser(r *http.Request) *database.User {
	user, ok := r.Context().Value(authenticatedUserContextKey).(*database.User)
	if !ok {
		return nil
	}

	return user
}

func contextSetGroup(r *http.Request, group *database.Group) *http.Request {
	ctx := context.WithValue(r.Context(), groupContextKey, group)
	return r.WithContext(ctx)
}

func contextGetGroup(r *http.Request) *database.Group {
	group, ok := r.Context().Value(groupContextKey).(*database.Group)
	if !ok {
		return nil
	}
	return group
}
