package websocket

import (
	"encoding/json"
	"fmt"
	"log"

	"brainbook-api/internal/response"
	t "brainbook-api/internal/time"
	"brainbook-api/internal/validator"
	"github.com/gorilla/websocket"
)

// SendMessageHandler will send out a message to all other participants in the chat
func SendMessageHandler(event Event, c *Client) error {
	log.Printf("SendMessageHandler called")
	// Marshal Payload into wanted format
	var chatevent SendMessageEvent
	if err := json.Unmarshal(event.Payload, &chatevent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}
	log.Printf("Unmarshaled message event: %+v", chatevent)

	// CRITICAL: Validate session token against database
	user, found, err := c.manager.DB.UserBySession(chatevent.SessionToken)
	if err != nil || !found {
		c.closeWithReason(websocket.ClosePolicyViolation, "Invalid session token")
		return nil
	}

	// Validate message content using validator package
	var v validator.Validator
	log.Printf("Validating message: length=%d, content='%s'", len(chatevent.Message), chatevent.Message)
	v.CheckField(validator.NotBlank(chatevent.Message), "message", "Message cannot be empty")
	v.CheckField(validator.MaxRunes(chatevent.Message, 250), "message", "Message is too long (maximum 250 characters)")
	log.Printf("Validation result: HasErrors=%v, FieldErrors=%+v", v.HasErrors(), v.FieldErrors)

	// Validate receiver ID
	v.CheckField(chatevent.ReceiverID > 0, "receiver_id", "Invalid receiver ID")

	if v.HasErrors() {
		// Send validation errors as WebSocket error with context
		for field, errorMsg := range v.FieldErrors {
			c.sendErrorEventWithContext("VALIDATION_ERROR", errorMsg, map[string]interface{}{
				"field":       field,
				"receiver_id": chatevent.ReceiverID,
			})
		}
		return nil
	}

	// Prevent users from sending messages to themselves
	if user.ID == chatevent.ReceiverID {
		c.sendErrorEventWithContext("SELF_MESSAGE", "Cannot send messages to yourself", map[string]interface{}{
			"receiver_id": chatevent.ReceiverID,
		})
		return nil
	}

	canMessage, err := c.manager.DB.CanUsersMessage(user.ID, chatevent.ReceiverID)
	if err != nil {
		return fmt.Errorf("failed to verify messaging permission: %v", err)
	}
	if !canMessage {
		c.sendErrorEventWithContext("MESSAGE_NOT_ALLOWED", "You cannot send messages to this user", map[string]interface{}{
			"receiver_id": chatevent.ReceiverID,
		})
		return nil
	}

	// Check if receiver is online
	log.Printf("Looking for receiver client with ID: %d", chatevent.ReceiverID)
	receiverClient := c.manager.getClientByUserID(chatevent.ReceiverID)
	if receiverClient == nil {
		log.Printf("Receiver client %d not found, sending RECEIVER_OFFLINE error", chatevent.ReceiverID)
		c.sendErrorEventWithContext("RECEIVER_OFFLINE", "Recipient is not online", map[string]interface{}{
			"receiver_id": chatevent.ReceiverID,
		})
		return nil
	}
	log.Printf("Found receiver client %d, proceeding with message send", chatevent.ReceiverID)

	// Generate single timestamp for consistency between database and client
	currentDateTime := t.CurrentTime()

	conversation, exists, err := c.manager.DB.ConversationByUserIDs(user.ID, chatevent.ReceiverID)
	if err != nil {
		return fmt.Errorf("failed to fetch conversation: %v", err)
	}

	var conversationID int
	if exists {
		conversationID = conversation.ID
	} else {
		conversationID, err = c.manager.DB.InsertConversation(user.ID, chatevent.ReceiverID, currentDateTime, currentDateTime)
		if err != nil {
			return fmt.Errorf("failed to create conversation: %v", err)
		}
	}

	// Save message to database
	_, err = c.manager.DB.InsertMessage(conversationID, user.ID, chatevent.Message, currentDateTime)
	if err != nil {
		return fmt.Errorf("failed to save message: %v", err)
	}

	_ = c.manager.DB.UpdateConversationLastMessageTime(conversationID, currentDateTime)

	// Prepare outgoing message
	var broadMessage ReceiveMessageEvent
	broadMessage.Message = chatevent.Message
	broadMessage.SenderID = user.ID
	broadMessage.ReceiverID = chatevent.ReceiverID
	broadMessage.SentAt = currentDateTime

	data, err := response.EncodeJSON(broadMessage)
	if err != nil {
		return fmt.Errorf("failed to encode broadcast message: %v", err)
	}

	// Create event
	var outgoingEvent Event
	outgoingEvent.Payload = data
	outgoingEvent.Type = EventReceiveMessage

	// Send to receiver
	receiverClient.egress <- outgoingEvent

	// Send confirmation back to sender
	c.egress <- outgoingEvent

	c.manager.CreateAndPushNotification(chatevent.ReceiverID, "direct_message", map[string]interface{}{
		"sender_id":       user.ID,
		"conversation_id": conversationID,
		"message":         chatevent.Message,
	})

	return nil
}

func SendGroupMessageHandler(event Event, c *Client) error {
	var payload SendGroupMessageEvent
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	user, found, err := c.manager.DB.UserBySession(payload.SessionToken)
	if err != nil || !found {
		c.closeWithReason(websocket.ClosePolicyViolation, "Invalid session token")
		return nil
	}

	var v validator.Validator
	v.CheckField(validator.NotBlank(payload.Message), "message", "Message cannot be empty")
	v.CheckField(payload.GroupID > 0, "group_id", "Invalid group ID")
	if v.HasErrors() {
		for field, msg := range v.FieldErrors {
			c.sendErrorEventWithContext("VALIDATION_ERROR", msg, map[string]interface{}{"field": field})
		}
		return nil
	}

	isMember, err := c.manager.DB.IsGroupMember(payload.GroupID, user.ID)
	if err != nil {
		return err
	}
	if !isMember {
		c.sendErrorEvent("GROUP_MESSAGE_FORBIDDEN", "You are not a member of this group")
		return nil
	}

	currentTime := t.CurrentTime()
	_, err = c.manager.DB.InsertGroupMessage(payload.GroupID, user.ID, payload.Message, currentTime)
	if err != nil {
		return err
	}

	message := ReceiveGroupMessageEvent{
		Message:  payload.Message,
		SenderID: user.ID,
		GroupID:  payload.GroupID,
		SentAt:   currentTime,
	}

	data, err := response.EncodeJSON(message)
	if err != nil {
		return err
	}

	eventOut := Event{Type: EventReceiveGroupMessage, Payload: data}
	members, err := c.manager.DB.GroupMembersByGroupID(payload.GroupID)
	if err != nil {
		return err
	}

	for _, member := range members {
		client := c.manager.getClientByUserID(member.ID)
		if client != nil {
			select {
			case client.egress <- eventOut:
			default:
				log.Printf("client %d egress full, dropping group message", client.userID)
			}
		} else {
			if member.ID == user.ID {
				continue
			}
			c.manager.CreateAndPushNotification(member.ID, "group_message", map[string]interface{}{
				"group_id":  payload.GroupID,
				"sender_id": user.ID,
				"message":   payload.Message,
			})
		}
	}

	return nil
}

// SendTypingHandler handles typing status updates between users
func SendTypingHandler(event Event, c *Client) error {
	// Marshal Payload into wanted format
	var typingEvent SendTypingEvent
	if err := json.Unmarshal(event.Payload, &typingEvent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	// CRITICAL: Validate session token against database
	user, found, err := c.manager.DB.UserBySession(typingEvent.SessionToken)
	if err != nil || !found {
		c.closeWithReason(websocket.ClosePolicyViolation, "Invalid session token")
		return nil
	}

	// Prevent users from sending typing status to themselves
	if user.ID == typingEvent.ReceiverID {
		c.sendErrorEvent("SELF_TYPING", "Cannot send typing status to yourself")
		return nil
	}

	// Check if receiver is online
	receiverClient := c.manager.getClientByUserID(typingEvent.ReceiverID)
	if receiverClient == nil {
		// Silently ignore if receiver is offline (no error needed for typing)
		return nil
	}

	// Prepare outgoing typing event
	var broadTyping NewTypingEvent
	broadTyping.SenderID = user.ID
	broadTyping.ReceiverID = typingEvent.ReceiverID
	broadTyping.IsTyping = typingEvent.IsTyping

	data, err := response.EncodeJSON(broadTyping)
	if err != nil {
		return fmt.Errorf("failed to encode typing event: %v", err)
	}

	// Create event
	var outgoingEvent Event
	outgoingEvent.Payload = data
	outgoingEvent.Type = EventNewTyping

	// Send to receiver only (sender doesn't need to know their own typing status)
	receiverClient.egress <- outgoingEvent

	return nil
}
