package api

import (
	"encoding/json"
	"log"
)

const (
	NotificationTypeDirectMessage = "direct_message"
	NotificationTypeGroupMessage  = "group_message"
	NotificationTypeGroupInvite   = "group_invite"
	NotificationTypeGroupJoin     = "group_join_request"
	NotificationTypeGroupEvent    = "group_event"
	NotificationTypeFollowRequest = "follow_request"
)

func (app *Application) notifyUser(userID int, notifType string, payload map[string]interface{}) {
	if app.DB == nil {
		return
	}

	var payloadBytes []byte
	var err error
	if payload != nil {
		payloadBytes, err = json.Marshal(payload)
		if err != nil {
			log.Printf("notifyUser marshal payload error: %v", err)
			return
		}
	}

	notif, err := app.DB.CreateNotification(userID, notifType, payloadBytes)
	if err != nil {
		log.Printf("notifyUser DB error: %v", err)
		return
	}

	if app.WSManager != nil {
		app.WSManager.PushNotification(notif)
	}
}
