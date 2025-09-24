package database

import (
	"context"
	"time"
)

type Post struct {
	Id           int       `db:"id" json:"id"`
	FirstName    string    `db:"f_name" json:"f_name"`
	LastName     string    `db:"l_name" json:"l_name"`
	Avatar       []byte    `db:"avatar" json:"avatar"`
	Content      string    `db:"content" json:"content"`
	File         []byte    `db:"file" json:"file"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	CommentCount int       `db:"comment_count" json:"comment_count"`
	Comments     []Comment `json:"comments"`
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

func (db *DB) InsertPost(content string, file []byte, currentDateTime string, userID int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
    INSERT INTO post (content, file, created_at, user_id) 
    VALUES ($1, $2, $3, $4)`

	result, err := db.ExecContext(ctx, query, content, file, currentDateTime, userID)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}
