package database

import (
	"context"
	"time"
)

type GroupPost struct {
	ID           int       `db:"id" json:"id"`
	GroupID      int       `db:"group_id" json:"group_id"`
	Content      string    `db:"content" json:"content"`
	File         []byte    `db:"file" json:"file,omitempty"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	CommentCount int       `db:"comment_count" json:"comment_count"`
	Comments     []Comment `json:"comments"`

	UserSummary
}

func (db *DB) InsertGroupPost(content string, file []byte, currentDateTime string, userID int, groupID int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		INSERT INTO group_posts (user_id, group_id, content, file, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	result, err := db.ExecContext(ctx, query, userID, groupID, content, file, currentDateTime)
	if err != nil {
		return 0, err
	}

	postID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(postID), nil
}

func (db *DB) GetGroupPosts(groupID int) ([]GroupPost, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
	SELECT 
		p.id,
		p.group_id,
		u.id as user_id,
		u.f_name,
		u.l_name,
		u.avatar,
		p.content,
		p.file,
		p.created_at,
		COALESCE(COUNT(gpc.id), 0) as comment_count
	FROM group_posts p
	JOIN user u ON p.user_id = u.id
	LEFT JOIN group_post_comments gpc ON gpc.group_post_id = p.id
	WHERE p.group_id = $1
	GROUP BY p.id, p.group_id, u.id, u.f_name, u.l_name, u.avatar, p.content, p.file, p.created_at
	ORDER BY p.created_at DESC
	`

	var posts []GroupPost
	err := db.SelectContext(ctx, &posts, query, groupID)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func (db *DB) GetGroupPostByID(groupPostID int) (*GroupPost, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		SELECT 
			gp.id,
			gp.group_id,
			u.id as user_id,
			u.f_name,
			u.l_name,
			u.avatar,
			gp.content,
			gp.file,
			gp.created_at,
			COALESCE(COUNT(gpc.id), 0) AS comment_count
		FROM group_posts AS gp
		JOIN user AS u ON gp.user_id = u.id
		LEFT JOIN group_post_comments AS gpc ON gpc.group_post_id = gp.id
		WHERE gp.id = $1
		GROUP BY 
			gp.id, gp.group_id, u.id, u.f_name, u.l_name, u.avatar, gp.content, gp.file, gp.created_at
	`

	var groupPost GroupPost
	err := db.GetContext(ctx, &groupPost, query, groupPostID)
	if err != nil {
		return nil, err
	}

	return &groupPost, nil
}
