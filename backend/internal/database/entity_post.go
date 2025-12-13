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
	Visibility   string    `db:"visibility" json:"visibility"`

	UserSummary
}

func (db *DB) InsertPost(userID int, content string, file []byte, visibility string, currentDateTime string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
	    INSERT INTO post (user_id, content, file, visibility, created_at) 
	    VALUES ($1, $2, $3, $4, $5)`

	result, err := db.ExecContext(ctx, query, userID, content, file, visibility, currentDateTime)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// AddPostViewers inserts allowed viewers for a limited-visibility post.
func (db *DB) AddPostViewers(postID int, userIDs []int) error {
	if len(userIDs) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		INSERT OR IGNORE INTO post_user_can_view (post_id, user_id)
		VALUES ($1, $2)
	`

	for _, uid := range userIDs {
		if _, err := db.ExecContext(ctx, query, postID, uid); err != nil {
			return err
		}
	}

	return nil
}

// CanUserViewPost returns true if viewerID can view the given postID.
func (db *DB) CanUserViewPost(viewerID, postID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var ownerID int
	var visibility string
	if err := db.QueryRowContext(ctx, "SELECT user_id, visibility FROM post WHERE id = $1", postID).Scan(&ownerID, &visibility); err != nil {
		return false, err
	}

	// Owner always sees their own post.
	if viewerID == ownerID {
		return true, nil
	}

	// Public posts visible to all.
	if visibility == "public" {
		return true, nil
	}

	// Followers-only posts require accepted follow.
	if visibility == "private" {
		return db.IsFollowing(viewerID, ownerID)
	}

	// Limited posts require explicit allow list.
	if visibility == "limited" {
		var count int
		query := `SELECT COUNT(*) FROM post_user_can_view WHERE post_id = $1 AND user_id = $2`
		if err := db.GetContext(ctx, &count, query, postID, viewerID); err != nil {
			return false, err
		}
		return count > 0, nil
	}

	return false, nil
}

// return posts by a target user that the context user can view
func (db *DB) PostsVisibleFromUser(viewerID, targetUserID int) ([]Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var posts []Post

	query := `
		SELECT 
			p.id, u.f_name, u.l_name, u.avatar,
			p.content, p.file, p.created_at, p.visibility,
			COALESCE(COUNT(c.id), 0) AS comment_count
		FROM post p
		JOIN user u ON p.user_id = u.id
		LEFT JOIN post_comment c ON p.id = c.post_id
		WHERE 
			p.user_id = $2
			AND (
				-- Viewer is the same as the target → can see everything
				$1 = $2

				-- Or public posts
				OR p.visibility = 'public'

				-- Or private posts (if viewer follows target)
				OR (
					p.visibility = 'private'
					AND EXISTS (
						SELECT 1
						FROM follow_request f
						WHERE f.requester_id = $1
						  AND f.target_id = $2
						  AND f.status = 'accepted'
					)
				)

				-- Or limited posts (if viewer is explicitly allowed)
				OR (
					p.visibility = 'limited'
					AND EXISTS (
						SELECT 1
						FROM post_user_can_view pcv
						WHERE pcv.post_id = p.id
						  AND pcv.user_id = $1
					)
				)
			)
		GROUP BY p.id, u.f_name, u.l_name, u.avatar, p.content, p.file, p.created_at, p.visibility
		ORDER BY p.created_at DESC;
	`

	err := db.SelectContext(ctx, &posts, query, viewerID, targetUserID)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

// TODO: modify to return all posts with a condition for private ones
func (db *DB) PrivatePostsByUserID(userID int) ([]Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var posts []Post

	query := `
		SELECT p.id, u.f_name, u.l_name, u.avatar, p.content, p.file, p.created_at, p.visibility,
		COALESCE(COUNT(c.id), 0) as comment_count
		FROM post p
		JOIN user u ON p.user_id = u.id
		LEFT JOIN post_comment c ON p.id = c.post_id
		WHERE p.user_id = $1 AND p.visibility = 'private'
		GROUP BY p.id, u.f_name, u.l_name, u.avatar, p.content, p.file, p.created_at, p.visibility
		ORDER BY p.created_at DESC`

	err := db.SelectContext(ctx, &posts, query, userID)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

// Retrieves all private & public posts a user can view
func (db *DB) AllPostsByUserID(userID int) ([]Post, error) {

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var posts []Post

	query := ` SELECT 
	    p.id,
	    u.f_name,
	    u.l_name,
	    u.avatar,
	    p.content,
	    p.file,
		    p.created_at,
		    p.visibility,
	    COALESCE(COUNT(c.id), 0) AS comment_count
	FROM post p
	JOIN user u 
	    ON p.user_id = u.id
	LEFT JOIN post_comment c 
	    ON p.id = c.post_id
	WHERE 
			(
				p.visibility = 'public'

				OR (
					p.visibility = 'private'
					AND EXISTS (
						SELECT 1 
						FROM follow_request f
						WHERE f.requester_id = $1       -- current user
						AND f.target_id = p.user_id   -- poster
						AND f.status = 'accepted'
					)
				)

				OR (
					p.visibility = 'limited'
					AND EXISTS (
						SELECT 1
						FROM post_user_can_view pcv
						WHERE pcv.post_id = p.id
						AND pcv.user_id = $1
					)
				)

				OR p.user_id = $1
			)
	GROUP BY 
	    p.id, u.f_name, u.l_name, u.avatar, p.content, p.file, p.created_at, p.visibility
	ORDER BY 
	    p.created_at DESC`

	err := db.SelectContext(ctx, &posts, query, userID)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

// Retrieves all limited posts visible to the given user
func (db *DB) LimitedPostsByUserID(userID int) ([]Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var posts []Post

	query := `
		SELECT 
			p.id,
			u.f_name,
			u.l_name,
			u.avatar,
			p.content,
			p.file,
			p.created_at,
			p.visibility,
			COALESCE(COUNT(c.id), 0) AS comment_count
		FROM post p
		JOIN user u 
			ON p.user_id = u.id
		LEFT JOIN post_comment c 
			ON p.id = c.post_id
		JOIN post_user_can_view pcv 
			ON pcv.post_id = p.id
		WHERE 
			p.visibility = 'limited'
			AND pcv.user_id = $1  
		GROUP BY 
			p.id, u.f_name, u.l_name, u.avatar, p.content, p.file, p.created_at, p.visibility
		ORDER BY 
			p.created_at DESC
	`

	err := db.SelectContext(ctx, &posts, query, userID)
	if err != nil {
		return nil, err
	}

	return posts, nil
}
