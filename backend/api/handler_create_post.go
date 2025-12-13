package api

import (
	"fmt"
	"net/http"
	"strings"

	"brainbook-api/internal/request"
	"brainbook-api/internal/response"
	t "brainbook-api/internal/time"
	"brainbook-api/internal/validator"
)

func (app *Application) createPost(w http.ResponseWriter, r *http.Request) {
	// Define the input structure to decode JSON
	var input struct {
		Content        string              `json:"content"`
		File           []byte              `json:"file"`
		Visibility     string              `json:"visibility"`
		AllowedUserIDs []int               `json:"allowed_user_ids"`
		Validator      validator.Validator `json:"-"`
	}

	// Decode the JSON request body
	err := request.DecodeJSON(w, r, &input)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	// Get the authenticated user from context
	user := contextGetAuthenticatedUser(r)

	input.Validator.CheckField(validator.NotBlank(input.Content), "post-content", "Content must not be empty")
	input.Validator.CheckField(validator.MaxRunes(input.Content, 500), "post-content", "Content must not exceed 500 characters")
	if len(input.File) > 0 {
		input.Validator.CheckField(len(input.File) <= 10_000_000, "file", "File size must be 10MB or less")
	}

	visibility := strings.ToLower(strings.TrimSpace(input.Visibility))
	if visibility == "" {
		visibility = "public"
	}

	var dbVisibility string
	switch visibility {
	case "public":
		dbVisibility = "public"
	case "almost_private", "followers", "followers_only":
		dbVisibility = "private" // followers-only
	case "private":
		dbVisibility = "limited" // selected followers
	default:
		input.Validator.AddFieldError("visibility", "Visibility must be one of public, almost_private, or private")
	}

	// Validate allow-list when using selected followers privacy.
	if dbVisibility == "limited" {
		if len(input.AllowedUserIDs) == 0 {
			input.Validator.AddFieldError("allowed_user_ids", "Provide at least one allowed follower for private posts")
		} else {
			for _, uid := range input.AllowedUserIDs {
				if uid <= 0 {
					input.Validator.AddFieldError("allowed_user_ids", "All allowed user IDs must be positive")
					break
				}
				isFollower, err := app.DB.IsFollowing(uid, user.ID)
				if err != nil {
					app.serverError(w, r, err)
					return
				}
				if !isFollower {
					input.Validator.AddFieldError("allowed_user_ids", fmt.Sprintf("User %d is not a follower", uid))
					break
				}
			}
		}
	}

	if input.Validator.HasErrors() {
		app.failedValidation(w, r, input.Validator)
		return
	}

	currentDateTime := t.CurrentTime()

	// Insert the post into the database
	postID, err := app.DB.InsertPost(user.ID, input.Content, input.File, dbVisibility, currentDateTime)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Persist allow-list for private posts restricted to selected followers
	if dbVisibility == "limited" {
		if err := app.DB.AddPostViewers(postID, input.AllowedUserIDs); err != nil {
			app.serverError(w, r, err)
			return
		}
	}

	// Respond with the created post
	responseData := map[string]any{
		"post_id":        postID,
		"user_full_name": user.FullName(),
		"user_avatar":    user.Avatar,
		"content":        input.Content,
		"file":           input.File,
		"visibility":     dbVisibility,
		"created_at":     currentDateTime,
	}

	err = response.JSON(w, http.StatusCreated, responseData)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
}
