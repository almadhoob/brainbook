package database

import (
	"context"
	"time"
)

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
	SELECT id, user_id, title, description, time, group_id
	FROM event
	WHERE group_id = $1
	ORDER BY time ASC
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
