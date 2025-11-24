package database

import (
	"context"
	"time"
)

type Comment struct {
	Content   string    `db:"content" json:"content"`
	File      []byte    `db:"file" json:"file"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`

	UserSummary
}

func (db DB) InsertComment(postID int, userID int, content string, file []byte, createdAt string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
    INSERT INTO post_comment (post_id, user_id, content, file, created_at)
    VALUES ($1, $2, $3, $4, $5)`

	result, err := db.ExecContext(ctx, query, postID, userID, content, file, createdAt)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), err
}

func (db *DB) CommentsForPost(postID int) ([]Comment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var comments []Comment

	// SUGGESTION: Listing selected fields like this is worth considering (better readability).
	query := `
	SELECT u.f_name,
		u.l_name,
		u.avatar,
		c.content,
		c.file,
		c.created_at
	FROM post_comment c
	JOIN user u ON c.user_id = u.id
	WHERE c.post_id = $1
	ORDER BY c.created_at ASC`

	err := db.SelectContext(ctx, &comments, query, postID)
	if err != nil {
		return nil, err
	}

	return comments, nil
}
