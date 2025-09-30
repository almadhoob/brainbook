package database

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Conversation struct {
	ID                int       `db:"id" json:"id"`
	User1ID           int       `db:"user1_id" json:"user1_id"`
	User2ID           int       `db:"user2_id" json:"user2_id"`
	Last_message_time time.Time `db:"last_message_time" json:"last_message_time"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
}

type Message struct {
	SenderID  int       `db:"sender_id" json:"sender_id"`
	Content   string    `db:"content" json:"content"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

func (db *DB) ConversationByUserIDs(user1ID, user2ID int) (*Conversation, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var conversation Conversation

	query := `
	SELECT * FROM conversation c
	WHERE (c.user1_id = $1 AND c.user2_id  = $2) 
       OR (c.user1_id  = $2 AND c.user2_id  = $1)`

	err := db.GetContext(ctx, &conversation, query, user1ID, user2ID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}

	return &conversation, true, err
}

func (db *DB) InsertMessage(senderid int, content string, currentDateTime string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
    INSERT INTO conversation_message (sender_id, message, created_at)
    VALUES ($1, $2, $3)`

	result, err := db.ExecContext(ctx, query, senderid, content, currentDateTime)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), err
}

func (db *DB) InsertConversation(user1ID, user2ID int, lastMessageTime, currentDateTime string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
    INSERT INTO conversation (user1_id, user2_id, last_message_time, created_at)
    VALUES ($1, $2, $3, $4)`

	result, err := db.ExecContext(ctx, query, user1ID, user2ID, lastMessageTime, currentDateTime)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), err
}

func (db *DB) UpdateConversationLastMessageTime(lastMessageTime string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `UPDATE conversation SET last_message_time = $1`

	_, err := db.ExecContext(ctx, query, lastMessageTime)
	return err
}

func (db *DB) PaginatedConversationMessagesByUserIDs(contextUserID, targetUserID, offset, limit int) ([]Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var messages []Message

	query := `
    SELECT m.sender_id, m.content, m.created_at
    FROM conversation_message m
	JOIN user u ON m.sender_id = u.id
	WHERE m.sender_id = $1 OR m.sender_id = $2
    ORDER BY m.created_at DESC
    LIMIT $3 OFFSET $4`

	err := db.SelectContext(ctx, &messages, query, contextUserID, targetUserID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Reverse the order to show oldest first (for chat display)
	for i := 0; i < len(messages)/2; i++ {
		j := len(messages) - 1 - i
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// GetMessageCount returns total number of messages between two users
func (db *DB) MessageCount(conversationID int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var count int
	query := `
    SELECT COUNT(*) 
    FROM message 
	JOIN private_conversation c ON m.conversation_id = c.id
	WHERE c.id = $1`

	err := db.GetContext(ctx, &count, query, conversationID)
	return count, err
}
