package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"brainbook-api/internal/response"
	v "brainbook-api/internal/validator"
)

type MessageHistoryResponse struct {
	SenderID  int    `json:"sender_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

func (app *Application) getConversation(w http.ResponseWriter, r *http.Request) {
	var validator v.Validator

	contextUser := contextGetAuthenticatedUser(r)
	pathUserID := r.PathValue("id")

	targetUserID, err := parseStringID(pathUserID)
	if err != nil {
		app.badRequest(w, r, fmt.Errorf("Invalid conversation ID: %s", pathUserID))
		return
	}

	conversation, exists, err := app.DB.ConversationByUserIDs(contextUser.ID, targetUserID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if !exists {
		// Respond with an empty slice when no conversation exists.
		responseData := map[string]interface{}{
			"messages":       []MessageHistoryResponse{},
			"target_user_id": targetUserID,
		}

		err = response.JSON(w, http.StatusOK, responseData)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
		return
	}

	// Parse pagination parameters
	pageParameter := r.URL.Query().Get("page")
	limitParameter := r.URL.Query().Get("limit")

	cleanPageParameter := strings.TrimSpace(pageParameter)
	cleanLimitParameter := strings.TrimSpace(limitParameter)

	// Default values
	var page = 1
	var limit = 10

	// Parse and validate page parameter
	if cleanPageParameter != "" {
		parsedPage, err := strconv.Atoi(cleanPageParameter)
		if err != nil {
			validator.AddError("Invalid page parameter value")
		} else {
			validator.Check(v.MinInt(parsedPage, 1), "Page value must be at least 1")
			validator.Check(v.MaxInt(parsedPage, 10), "Page value cannot exceed 1000")
			page = parsedPage
		}
	}

	// Parse and validate limit parameter
	if cleanLimitParameter != "" {
		parsedLimit, err := strconv.Atoi(cleanLimitParameter)
		if err != nil {
			validator.AddError("Invalid limit paramete value")
		} else {
			validator.Check(v.MinInt(parsedLimit, 1), "Limit value must be at least 1")
			validator.Check(v.MaxInt(parsedLimit, 20), "Limit value cannot exceed 100")
			limit = parsedLimit
		}
	}

	offset := (page - 1) * limit

	messages, err := app.DB.PaginatedConversationMessages(conversation.ID, offset, limit)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	totalMessageCount, err := app.DB.MessageCount(conversation.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	var messageHistory []MessageHistoryResponse
	for _, Message := range messages {
		messageHistory = append(messageHistory, MessageHistoryResponse{
			SenderID:  Message.SenderID,
			Content:   Message.Content,
			CreatedAt: Message.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	// Calculate pagination metadata
	totalPages := (totalMessageCount + limit - 1) / limit
	hasNext := page < totalPages
	hasPrevious := page > 1

	responseData := map[string]interface{}{
		"messages": messageHistory,
		"pagination": map[string]interface{}{
			"current_page": page,
			"total_pages":  totalPages,
			"total_count":  totalMessageCount,
			"limit":        limit,
			"has_next":     hasNext,
			"has_previous": hasPrevious,
		},
		"target_user_id": targetUserID,
	}

	err = response.JSON(w, http.StatusOK, responseData)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
}
