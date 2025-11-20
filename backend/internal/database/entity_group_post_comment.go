package database

import (
	"context"
	"time"
)

type GroupPostComment struct {
	ID        int       `db:"id" json:"id"`
	File      []byte    `db:"file" json:"file,omitempty"`
	Content   string    `db:"content" json:"content"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`

	UserSummary
}

func (db *DB) InsertGroupPostComment(content string, file []byte, currentDateTime string, groupPostID int, userID int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		INSERT INTO group_post_comments (group_post_id, user_id, content, file, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	result, err := db.ExecContext(ctx, query, groupPostID, userID, content, file, currentDateTime)
	if err != nil {
		return 0, err
	}

	commentID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(commentID), nil
}

func (db *DB) GetCommentsForGroupPost(groupPostID int) ([]GroupPostComment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		SELECT 
			c.id,
			u.id as user_id,
			u.f_name,
			u.l_name,
			u.avatar,
			c.content,
			c.file,
			c.created_at
		FROM group_post_comments AS c
		JOIN user AS u ON c.user_id = u.id
		WHERE c.group_post_id = $1
		ORDER BY c.created_at ASC
	`

	var comments []GroupPostComment
	err := db.SelectContext(ctx, &comments, query, groupPostID)
	if err != nil {
		return nil, err
	}

	return comments, nil
}
