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

func (db *DB) InsertGroup(ownerID int, title string, description string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	query := `
		INSERT INTO groups (owner_id, title, description, created_at)
		VALUES ($1, $2, $3, $4)
	`
	result, err := db.ExecContext(ctx, query, ownerID, title, description, time.Now())
	if err != nil {
		return 0, err
	}

	groupID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(groupID), nil
}

func (db *DB) AllGroups() ([]Group, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `SELECT * FROM groups`

	var groups []Group
	err := db.SelectContext(ctx, &groups, query)
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (db *DB) GroupByID(groupID int) (*Group, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `SELECT * FROM groups WHERE id = $1`

	var group Group
	err := db.GetContext(ctx, &group, query, groupID)
	if err != nil {
		return nil, err
	}

	return &group, nil
}

func (db *DB) GroupsByUserID(userID int) ([]Group, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var groups []Group

	query := `
		SELECT g.id, g.owner_id, g.title, g.description, g.created_at
		FROM groups AS g
		JOIN group_members AS gm ON gm.group_id = g.id
		WHERE gm.user_id = $1
		ORDER BY g.title ASC`

	err := db.SelectContext(ctx, &groups, query, userID)
	if err != nil {
		return nil, err
	}

	return groups, nil
}
