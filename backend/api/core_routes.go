package api

import (
	"net/http"
)

func (app *Application) routes() http.Handler {
	// Create route registry
	registry := NewRouteRegistry()

	// Public routes (no authentication required)
	registry.HandleFunc("/js/", http.StripPrefix("/js/", app.neuteredFileHandler("./static/js/")).ServeHTTP).
		HandleFunc("/css/", http.StripPrefix("/css/", app.neuteredFileHandler("./static/css/")).ServeHTTP).
		HandleFunc("/images/", http.StripPrefix("/images/", app.neuteredFileHandler("./static/images/")).ServeHTTP).
		GetMethod("/v1/status", app.status).
		GetMethod("/v1/404", app.notFound).
		PostMethod("/v1/login", app.createAuthenticationToken).
		PostMethod("/v1/register", app.createUser)

	// Protected routes (authentication required)
	registry.GetMethod("/protected/ws", app.ServeWebSocket).
		GetMethod("/protected/v1/profile/user/{id}", app.getUserProfile).
		GetMethod("/protected/v1/user-list", app.getUserList).
		GetMethod("/protected/v1/private-messages/user/{id}", app.getConversation).
		GetMethod("/protected/v1/posts", app.getPosts).
		GetMethod("/protected/v1/comments", app.getPostComments).
		GetMethod("/protected/v1/notifications", app.getNotifications).
		GetMethod("/protected/v1/groups", app.getGroups).
		GetMethod("/protected/v1/user/groups", app.userGroups).
		GetMethod("/protected/v1/groups/{group_id}", app.withGroup(app.groupDetails)).
		GetMethod("/protected/v1/groups/{group_id}/members", app.requireGroupMember(app.getMembers)).
		GetMethod("/protected/v1/groups/{group_id}/posts", app.withGroup(app.groupPosts)).
		GetMethod("/protected/v1/groups/{group_id}/messages", app.requireGroupMember(app.getGroupMessages)).
		GetMethod("/protected/v1/groups/{group_id}/events", app.requireGroupMember(app.listGroupEvents)).
		GetMethod("/protected/v1/groups/{group_id}/posts/{post_id}/comments", app.requireGroupMember(app.getGroupPostComments)).
		PostMethod("/protected/v1/posts", app.createPost).
		PostMethod("/protected/v1/comments", app.createComment).
		PostMethod("/protected/v1/logout", app.logout).
		PostMethod("/protected/v1/profile/update", app.updateProfile).
		PostMethod("/protected/v1/notifications/{notification_id}/read", app.markNotificationRead).
		PostMethod("/protected/v1/users/{user_id}/follow", app.sendFollowRequest).
		PostMethod("/protected/v1/follow-requests/{request_id}", app.respondFollowRequest).
		PostMethod("/protected/v1/groups/{group_id}/create", app.requireGroupMember(app.groupPostCreate)).
		PostMethod("/protected/v1/groups/{group_id}/posts/{post_id}/comments", app.requireGroupMember(app.createGroupPostComment)).
		PostMethod("/protected/v1/groups/{group_id}/events", app.requireGroupOwner(app.createGroupEvent)).
		PostMethod("/protected/v1/groups/{group_id}/join", app.withGroup(app.joinGroupRequest)).
		PostMethod("/protected/v1/groups/{group_id}/send", app.requireGroupOwner(app.SendGroupInvite))

	publicMux, guestMux, protectedMux := registry.GetMuxes()

	// Add catchAll to publicMux for SPA routes
	publicMux.HandleFunc("/", app.catchAll)

	// Create final mux to mount different route groups with different middleware
	finalMux := http.NewServeMux()

	// Mount public routes (no authentication) - these include static files and public API
	finalMux.Handle("/", publicMux)

	// Mount guest routes (optional authentication - guests allowed)
	finalMux.Handle("/guest/", http.StripPrefix("/guest", app.authenticate(guestMux)))

	// Mount protected routes (authentication required)
	finalMux.Handle("/protected/", http.StripPrefix("/protected", app.authenticate(app.authorize(protectedMux))))

	// Apply method validation to the entire final mux
	return app.logAccess(app.recoverPanic(registry.ValidateMethod()(finalMux)))
}
