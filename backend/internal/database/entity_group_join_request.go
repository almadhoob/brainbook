package database

import (
	"context"
	"database/sql"
	"errors"
)

type GroupJoinRequest struct {
	GroupID     int    `db:"group_id" json:"group_id"`
	RequestID   int    `db:"request_id" json:"request_id"`
	RequesterID int    `db:"requester_id" json:"requester_id"`
	TargetID    int    `db:"target_id" json:"target_id"`
	Status      string `db:"status" json:"status"`
	CreatedAt   string `db:"created_at" json:"created_at"`

	UserSummary
}

// GroupJoinRequestByID fetches a request by its ID.
func (db *DB) GroupJoinRequestByID(requestID int) (*GroupJoinRequest, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		SELECT gjr.id AS request_id, gjr.group_id, gjr.status, gjr.created_at,
		       u.id AS user_id, u.f_name, u.l_name, u.avatar,
		       gjr.requester_id, gjr.target_id
		FROM group_join_requests gjr
		JOIN user u ON u.id = gjr.requester_id
		WHERE gjr.id = $1
	`

	var req GroupJoinRequest
	if err := db.GetContext(ctx, &req, query, requestID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return &req, true, nil
}

func (db *DB) InsertJoinRequest(groupID int, requesterID int, targetID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		INSERT INTO group_join_requests (group_id, requester_id, target_id, status)
		SELECT $1, $2, $3, 'pending'
		WHERE NOT EXISTS (
			SELECT 1 FROM group_join_requests
			WHERE group_id = $1 AND requester_id = $2 AND target_id = $3 AND status = 'pending'
		)
	`

	_, err := db.ExecContext(ctx, query, groupID, requesterID, targetID)
	return err
}

func (db *DB) RequestExistsAndPending(groupID int, requesterID int, targetID int) (bool, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var status string
	query := `
        SELECT status
        FROM group_join_requests
        WHERE group_id = $1 AND requester_id = $2 AND target_id = $3
        LIMIT 1
    `
	err := db.GetContext(ctx, &status, query, groupID, requesterID, targetID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}
	return true, status == "pending", nil
}

func (db *DB) PendingJoinRequestsByGroupID(groupID int) ([]GroupJoinRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
	SELECT
		gjr.id AS request_id,
		gjr.group_id,
		gjr.requester_id,
		gjr.target_id,
		u.id as user_id,
		u.f_name,
		u.l_name,
		u.avatar,
		gjr.status,
		gjr.created_at	
	FROM group_join_requests AS gjr
		JOIN user AS u ON gjr.requester_id = u.id
		WHERE gjr.group_id = $1
  		AND gjr.status = 'pending'
		ORDER BY gjr.created_at ASC`

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
		UPDATE group_join_requests
		SET status = $1
		WHERE id = $2
	`

	_, err := db.ExecContext(ctx, query, newStatus, requestID)
	return err
}
