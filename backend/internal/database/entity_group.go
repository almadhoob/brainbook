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

func (db *DB) InsertGroup(ownerID int, name string, title string, description string, createdAt string) (int, error) {
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
