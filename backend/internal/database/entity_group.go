package database

import (
	"context"
	"time"
)

type Group struct {
	ID          int       `db:"id" json:"id"`
	OwnerID     int       `db:"owner_id" json:"owner_id"`
	Title       string    `db:"title" json:"title"`
	Description string    `db:"description" json:"description"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type GroupMember struct {
	Role     string `db:"role" json:"role"`
	JoinedAt string `db:"joined_at" json:"joined_at"`
	
	UserSummary
}

type GroupJoinRequest struct {
	GroupID   int    `db:"group_id" json:"group_id"`
	RequestID int    `db:"request_id" json:"request_id"`
	Status    string `db:"status" json:"status"`
	CreatedAt string `db:"created_at" json:"created_at"`

	UserSummary
}

type GroupEvent struct {
	ID                 int           `db:"id" json:"id"`
	Title              string        `db:"title" json:"title"`
	Description        string        `db:"description" json:"description"`
	Time               time.Time     `db:"time" json:"time"`
	GroupID            int           `db:"group_id" json:"group_id"`
	InterestedCount    int           `db:"interested" json:"interested,omitempty"`
	NotInterestedCount int           `db:"not_interested" json:"not_interested,omitempty"`
	Participants       []GroupMember `json:"participants,omitempty"`
}

type GroupPost struct {
	ID           int       `db:"id" json:"id"` //post id
	GroupID      int       `db:"group_id" json:"group_id"`
	Content      string    `db:"content" json:"content"`
	Image        []byte    `db:"image" json:"image,omitempty"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	CommentCount int       `db:"comment_count" json:"comment_count"`
	Comments     []Comment `json:"comments"`

    UserSummary
}

type GroupPostComment struct {
	ID        int       `db:"id" json:"id"`
	File      []byte    `db:"file" json:"file,omitempty"`
	Content   string    `db:"content" json:"content"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`

	UserSummary
}


func (db *DB) InsertGroup(ownerID int, title string, description string, createdAt string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	query := `
		INSERT INTO group (owner_id, title, description, created_at)
		VALUES ($1, $2, $3, $4)
	`
	result, err := db.ExecContext(ctx, query, ownerID, title, description, createdAt)
	if err != nil {
		return 0, err
	}

	groupID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(groupID), nil
}

func (db *DB) GetAllGroups() ([]Group, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `SELECT * FROM group `

	var groups []Group
	err := db.SelectContext(ctx, &groups, query)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (db *DB) AddGroupMember(groupID int, userID int, role string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		INSERT INTO group_member (group_id, user_id, role)
		VALUES ($1, $2, $3)
	`
	_, err := db.ExecContext(ctx, query, groupID, userID, role)
	return err
}

// func (db *DB) RemoveGroupMember(groupID int, userID int) error {
// 	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
// 	defer cancel()

// 	query := `
// 		DELETE FROM group_member
// 		WHERE group_id = $1 AND user_id = $2
// 	`
// 	_, err := db.ExecContext(ctx, query, groupID, userID)
// 	return err
// }

func (db *DB) GetGroupMembers(groupID int) ([]GroupMember, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		SELECT 
			u.id as user_id,
			u.f_name,
			u.l_name,
			u.avatar,
			gm.role,
			gm.joined_at
		FROM group_member AS gm
		JOIN user AS u ON gm.user_id = u.id
		WHERE gm.group_id = $1;
	`

	var members []GroupMember
	err := db.SelectContext(ctx, &members, query, groupID)
	if err != nil {
		return nil, err
	}

	return members, nil
}

func (db *DB) SendJoinRequest(groupID int, requesterID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		INSERT INTO group_join_request (group_id, requester_id, status)
		VALUES ($1, $2, 'pending')
		WHERE NOT EXISTS (
			SELECT 1 FROM group_join_request
			WHERE group_id = $1 AND requester_id = $2 AND status = 'pending'
		);
	`

	_, err := db.ExecContext(ctx, query, groupID, requesterID)
	return err
}

// GetPendingJoinRequests retrieves all pending join requests for a specific group
func (db *DB) GetPendingJoinRequests(groupID int) ([]GroupJoinRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
	SELECT
		gjr.id AS request_id,
		gjr.group_id,
		gjr.requester_id,
		u.f_name,
		u.l_name,
		u.avatar,
		gjr.status,
		gjr.created_at	
	FROM group_join_request AS gjr
		JOIN user AS u ON gjr.requester_id = u.id
		WHERE gjr.group_id = $1
  		AND gjr.status = 'pending'
		ORDER BY gjr.created_at ASC;`

	var requests []GroupJoinRequest
	err := db.SelectContext(ctx, &requests, query, groupID)
	if err != nil {
		return nil, err
	}

	return requests, nil
}
func (db *DB) UpdateJoinRequestStatus(requestID int, newStatus string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		UPDATE group_join_request
		SET status = $1
		WHERE id = $2;
	`

	_, err := db.ExecContext(ctx, query, newStatus, requestID)
	return err
}


func (db *DB) InsertGroupEvent(user_id int, title string, description string, eventTime string, groupID int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
	INSERT INTO event (user_id, title, description, time, group_id)
	VALUES ($1, $2, $3, $4, $5);
	`

	result, err := db.ExecContext(ctx, query, user_id, title, description, eventTime, groupID)
	if err != nil {
		return 0, err
	}

	eventID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(eventID), nil
}

func (db *DB) GetGroupEvents(groupID int) ([]GroupEvent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
	SELECT id, user_id, title, description, time, group_id
	FROM event
	WHERE group_id = $1
	ORDER BY time ASC;
	`

	var events []GroupEvent
	err := db.SelectContext(ctx, &events, query, groupID)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (db *DB) GetEventParticipants(eventID int) ([]GroupMember, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
	SELECT 
		u.id as user_id,
		u.f_name,
		u.l_name,
		u.avatar,
		gm.role,
		gm.joined_at
	FROM event_participation AS ep
	JOIN group_member AS gm ON ep.user_id = gm.user_id
	JOIN user AS u ON gm.user_id = u.id
	WHERE ep.event_id = $1;
	`

	var participants []GroupMember
	err := db.SelectContext(ctx, &participants, query, eventID)
	if err != nil {
		return nil, err
	}

	return participants, nil
}

//queries need adjusting

func (db *DB) InsertGroupPost(content string, image []byte, currentDateTime string, userID int, groupID int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		INSERT INTO group_post (user_id, group_id, content, file, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	result, err := db.ExecContext(ctx, query, userID, groupID, content, image, currentDateTime)
	if err != nil {
		return 0, err
	}

	postID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(postID), nil
}

//retrieve posts for a specific group along with user details and comment count
func (db *DB) GetGroupPosts(groupID int) ([]GroupPost, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
	SELECT 
		p.id, p.user_id, u.f_name, u.l_name, u.avatar, gm.role, p.group_id, p.content, p.image, p.created_at,
		COALESCE(COUNT(c.id), 0) as comment_count
	FROM group_post p
	JOIN user u ON p.user_id = u.id
	JOIN group_member gm ON p.user_id = gm.user_id AND p.group_id = gm.group_id
	LEFT JOIN comment c ON p.id = c.post_id
	WHERE p.group_id = $1
	GROUP BY p.id
	ORDER BY p.created_at DESC;
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
			gp.id AS group_post_id,
			gp.group_id,
			g.name AS group_name,
			gp.user_id,
			u.f_name,
			u.l_name,
			u.avatar,
			gp.content,
			gp.image,
			gp.created_at,
			COUNT(c.id) AS comment_count
		FROM group_post AS gp
		JOIN group AS g ON gp.group_id = g.id
		JOIN user AS u ON gp.user_id = u.id
		LEFT JOIN comment AS c ON c.group_post_id = gp.id
		WHERE gp.id = $1
		GROUP BY 
			gp.id, gp.group_id, g.name, gp.user_id, 
			u.f_name, u.l_name, u.avatar, gp.content, gp.image, gp.created_at;
	`

	var groupPost GroupPost
	err := db.GetContext(ctx, &groupPost, query, groupPostID)
	if err != nil {
		return nil, err
	}

	return &groupPost, nil
}



func (db *DB) InsertGroupPostComment(content string, image []byte, currentDateTime string, groupPostID int, userID int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		INSERT INTO group_post_comment (group_post_id, user_id, content, file, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	result, err := db.ExecContext(ctx, query, groupPostID, userID, content, image, currentDateTime)
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
			c.id AS comment_id,
			c.group_post_id,
			c.user_id,
			u.f_name,
			u.l_name,
			u.avatar,
			c.content,
			c.file,
			c.created_at
		FROM group_post_comment AS c
		JOIN user AS u ON c.user_id = u.id
		WHERE c.group_post_id = $1
		ORDER BY c.created_at ASC;
	`

	var comments []GroupPostComment
	err := db.SelectContext(ctx, &comments, query, groupPostID)
	if err != nil {
		return nil, err
	}

	return comments, nil
}