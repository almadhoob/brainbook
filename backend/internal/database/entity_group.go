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
			u.id,
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
		VALUES (?, ?, 'pending')
		WHERE NOT EXISTS (
			SELECT 1 FROM group_join_request 
			WHERE group_id = ? AND requester_id = ? AND status = 'pending'
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
		SET status = ?
		WHERE id = ?;
	`

	_, err := db.ExecContext(ctx, query, newStatus, requestID)
	return err
}
