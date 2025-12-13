package database

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type GroupEvent struct {
	ID                 int           `db:"id" json:"id"`
	UserID             int           `db:"user_id" json:"user_id"`
	Title              string        `db:"title" json:"title"`
	Description        string        `db:"description" json:"description"`
	Time               time.Time     `db:"time" json:"time"`
	GroupID            int           `db:"group_id" json:"group_id"`
	InterestedCount    int           `db:"interested" json:"interested,omitempty"`
	NotInterestedCount int           `db:"not_interested" json:"not_interested,omitempty"`
	Participants       []GroupMember `json:"participants,omitempty"`
}

// EventByID fetches event metadata.
func (db *DB) EventByID(eventID int) (*GroupEvent, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `SELECT id, user_id, title, description, time, group_id FROM event WHERE id = $1`
	var ev GroupEvent
	if err := db.GetContext(ctx, &ev, query, eventID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return &ev, true, nil
}

func (db *DB) InsertGroupEvent(userID int, title string, description string, eventTime string, groupID int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
	INSERT INTO event (user_id, title, description, time, group_id)
	VALUES ($1, $2, $3, $4, $5)
	`

	result, err := db.ExecContext(ctx, query, userID, title, description, eventTime, groupID)
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
	SELECT 
		e.id,
		e.user_id,
		e.title,
		e.description,
		e.time,
		e.group_id,
		COALESCE(SUM(CASE WHEN ehu.interested = 1 THEN 1 ELSE 0 END), 0) AS interested,
		COALESCE(SUM(CASE WHEN ehu.interested = 0 THEN 1 ELSE 0 END), 0) AS not_interested
	FROM event e
	LEFT JOIN event_has_user ehu ON ehu.event_id = e.id
	WHERE e.group_id = $1
	GROUP BY e.id, e.user_id, e.title, e.description, e.time, e.group_id
	ORDER BY e.time ASC
	`

	var events []GroupEvent
	err := db.SelectContext(ctx, &events, query, groupID)
	if err != nil {
		return nil, err
	}

	return events, nil
}

// UpsertEventRSVP records a going/not-going response for a user.
func (db *DB) UpsertEventRSVP(eventID int, userID int, going bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		INSERT INTO event_has_user (event_id, user_id, interested)
		VALUES ($1, $2, $3)
		ON CONFLICT(event_id, user_id) DO UPDATE SET interested = $3
	`

	_, err := db.ExecContext(ctx, query, eventID, userID, going)
	return err
}

func (db *DB) GetEventParticipants(eventID int) ([]GroupMember, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
	SELECT
    u.id       AS user_id,
    u.f_name,
    u.l_name,
    u.avatar,
    gm.role,
    gm.joined_at
FROM event_has_user AS ehu
JOIN user AS u
    ON ehu.user_id = u.id
LEFT JOIN group_members AS gm
    ON gm.user_id = u.id
   AND gm.group_id = (SELECT group_id FROM event WHERE id = $1)
WHERE ehu.event_id = $1;
	`

	var participants []GroupMember
	err := db.SelectContext(ctx, &participants, query, eventID)
	if err != nil {
		return nil, err
	}

	return participants, nil
}
