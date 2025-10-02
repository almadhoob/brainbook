package database

import (
	"context"
	"time"
)

type Post struct {
	ID           int       `db:"id" json:"id"`
	Content      string    `db:"content" json:"content"`
	File         []byte    `db:"file" json:"file"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	CommentCount int       `db:"comment_count" json:"comment_count"`
	Comments     []Comment `json:"comments"`

	UserSummary 
}

func (db *DB) PostsByUserID(userID int) (*Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var post Post

	query := `
    	SELECT p.id, u.f_name, u.l_name, u.avatar, p.content, p.file, p.created_at
		COALESCE(COUNT(c.id), 0) as comment_count
		FROM post p 
		WHERE p.user_id = $1            
    	JOIN user u ON p.user_id = u.id
		LEFT JOIN comment c ON p.id = c.post_id`

	err := db.GetContext(ctx, &post, query, userID)
	if err != nil {
		return nil, err
	}

	return &post, nil
}

func (db *DB) GetPosts() ([]Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var posts []Post

	query := `
        SELECT p.id, u.f_name, u.l_name, p.content, p.file, p.created_at, 
    	COALESCE(COUNT(c.id), 0) as comment_count
        FROM post p            
        JOIN user u ON p.user_id = u.id
        LEFT JOIN comment c ON p.id = c.post_id
        GROUP BY p.id, u.first_name, p.content, p.created_at
        ORDER BY p.created_at DESC`

	err := db.SelectContext(ctx, &posts, query)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func (db *DB) InsertPost(userID int, content string, file []byte, currentDateTime string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
    INSERT INTO post (user_id, content, file, created_at) 
    VALUES ($1, $2, $3, $4)`

	result, err := db.ExecContext(ctx, query, userID, content, file, currentDateTime)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}
