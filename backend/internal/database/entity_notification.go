package database

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// JSONPayload wraps json.RawMessage so it can be scanned from SQLite TEXT values
// while still serializing as raw JSON in API responses.
type JSONPayload json.RawMessage

// Scan implements sql.Scanner so TEXT columns can populate JSONPayload values.
func (p *JSONPayload) Scan(value interface{}) error {
	if p == nil {
		return fmt.Errorf("JSONPayload: Scan on nil pointer")
	}

	switch v := value.(type) {
	case nil:
		*p = nil
	case []byte:
		*p = JSONPayload(append([]byte(nil), v...))
	case string:
		*p = JSONPayload([]byte(v))
	default:
		return fmt.Errorf("JSONPayload: unsupported scan type %T", value)
	}
	return nil
}

// Value implements driver.Valuer so JSONPayload can be written back to SQLite.
func (p JSONPayload) Value() (driver.Value, error) {
	if p == nil {
		return nil, nil
	}
	return []byte(p), nil
}

// MarshalJSON preserves the raw JSON payload when encoding API responses.
func (p JSONPayload) MarshalJSON() ([]byte, error) {
	if p == nil {
		return []byte("null"), nil
	}
	return json.RawMessage(p).MarshalJSON()
}

type Notification struct {
	ID        int         `db:"id" json:"id"`
	UserID    int         `db:"user_id" json:"user_id"`
	Type      string      `db:"type" json:"type"`
	Payload   JSONPayload `db:"payload" json:"payload"`
	IsRead    bool        `db:"is_read" json:"is_read"`
	CreatedAt time.Time   `db:"created_at" json:"created_at"`
}

type notificationRecord struct {
	ID        int       `db:"id"`
	UserID    int       `db:"user_id"`
	Type      string    `db:"type"`
	Payload   []byte    `db:"payload"`
	IsRead    bool      `db:"is_read"`
	CreatedAt time.Time `db:"created_at"`
}

func (row notificationRecord) toNotification() Notification {
	return Notification{
		ID:        row.ID,
		UserID:    row.UserID,
		Type:      row.Type,
		Payload:   cloneJSONPayload(row.Payload),
		IsRead:    row.IsRead,
		CreatedAt: row.CreatedAt,
	}
}

func cloneJSONPayload(raw []byte) JSONPayload {
	if raw == nil {
		return nil
	}
	return JSONPayload(append([]byte(nil), raw...))
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

	var row notificationRecord
	query := `SELECT id, user_id, type, payload, is_read, created_at FROM notifications WHERE id = $1`
	if err := db.GetContext(ctx, &row, query, id); err != nil {
		return nil, err
	}
	notif := row.toNotification()
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

	var rows []notificationRecord
	if err := db.SelectContext(ctx, &rows, query, userID); err != nil {
		return nil, err
	}

	notifications := make([]Notification, len(rows))
	for i, row := range rows {
		notifications[i] = row.toNotification()
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
