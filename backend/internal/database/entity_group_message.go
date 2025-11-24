package database

import (
	"context"
	"time"
)

type GroupMessage struct {
	ID        int       `db:"id" json:"id"`
	GroupID   int       `db:"group_id" json:"group_id"`
	Content   string    `db:"content" json:"content"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`

	UserSummary
}

func (db *DB) InsertGroupMessage(groupID, senderID int, content, createdAt string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
        INSERT INTO group_messages (group_id, sender_id, content, created_at)
        VALUES ($1, $2, $3, $4)
    `

	result, err := db.ExecContext(ctx, query, groupID, senderID, content, createdAt)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (db *DB) GroupMessages(groupID, limit, offset int) ([]GroupMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	if limit <= 0 {
		limit = 25
	}

	query := `
        SELECT gm.id,
               gm.group_id,
               u.id as user_id,
               u.f_name,
               u.l_name,
               u.avatar,
               gm.content,
               gm.created_at
        FROM group_messages gm
        JOIN user u ON gm.sender_id = u.id
        WHERE gm.group_id = $1
        ORDER BY gm.created_at DESC
        LIMIT $2 OFFSET $3
    `

	var messages []GroupMessage
	if err := db.SelectContext(ctx, &messages, query, groupID, limit, offset); err != nil {
		return nil, err
	}

	// reverse to chronological order
	for i := 0; i < len(messages)/2; i++ {
		j := len(messages) - 1 - i
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}
