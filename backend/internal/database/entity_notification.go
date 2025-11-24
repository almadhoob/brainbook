package database

import (
	"context"
	"encoding/json"
	"time"
)

type Notification struct {
	ID        int             `db:"id" json:"id"`
	UserID    int             `db:"user_id" json:"user_id"`
	Type      string          `db:"type" json:"type"`
	Payload   json.RawMessage `db:"payload" json:"payload"`
	IsRead    bool            `db:"is_read" json:"is_read"`
	CreatedAt time.Time       `db:"created_at" json:"created_at"`
}

func (db *DB) CreateNotification(userID int, notifType string, payload []byte) (*Notification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
        INSERT INTO notifications (user_id, type, payload)
        VALUES ($1, $2, $3)
    `

	result, err := db.ExecContext(ctx, query, userID, notifType, payload)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return db.NotificationByID(int(id))
}

func (db *DB) NotificationByID(id int) (*Notification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var notif Notification
	query := `SELECT id, user_id, type, payload, is_read, created_at FROM notifications WHERE id = $1`
	if err := db.GetContext(ctx, &notif, query, id); err != nil {
		return nil, err
	}
	return &notif, nil
}

func (db *DB) NotificationsByUser(userID int, includeRead bool) ([]Notification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	baseQuery := `SELECT id, user_id, type, payload, is_read, created_at FROM notifications WHERE user_id = $1`
	var query string
	if includeRead {
		query = baseQuery + ` ORDER BY created_at DESC`
	} else {
		query = baseQuery + ` AND is_read = 0 ORDER BY created_at DESC`
	}

	var notifications []Notification
	if err := db.SelectContext(ctx, &notifications, query, userID); err != nil {
		return nil, err
	}
	return notifications, nil
}

func (db *DB) MarkNotificationRead(notificationID, userID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `UPDATE notifications SET is_read = 1 WHERE id = $1 AND user_id = $2`
	_, err := db.ExecContext(ctx, query, notificationID, userID)
	return err
}
