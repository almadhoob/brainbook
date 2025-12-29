package database

import "context"

type GroupMember struct {
	Role     string `db:"role" json:"role"`
	JoinedAt string `db:"joined_at" json:"joined_at"`

	UserSummary
}

func (db *DB) InsertGroupMember(groupID int, userID int, role string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		INSERT INTO group_members (group_id, user_id, role)
		VALUES ($1, $2, $3)
	`

	_, err := db.ExecContext(ctx, query, groupID, userID, role)
	return err
}

// InsertGroupMemberOrIgnore safely inserts membership without failing on duplicates.
func (db *DB) InsertGroupMemberOrIgnore(groupID int, userID int, role string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		INSERT OR IGNORE INTO group_members (group_id, user_id, role)
		VALUES ($1, $2, $3)
	`

	_, err := db.ExecContext(ctx, query, groupID, userID, role)
	return err
}

func (db *DB) IsGroupMember(groupID int, userID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var count int

	query := `
		SELECT COUNT(*)
		FROM group_members
		WHERE group_id = $1 AND user_id = $2
	`

	err := db.GetContext(ctx, &count, query, groupID, userID)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (db *DB) GroupMembersByGroupID(groupID int) ([]GroupMember, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		SELECT
			u.id as user_id,
			u.f_name,
			u.l_name,
			u.avatar,
			COALESCE(gm.role, 'owner') as role,
			COALESCE(gm.joined_at, g.created_at) as joined_at
		FROM groups AS g
		JOIN user AS u ON u.id = g.owner_id
		LEFT JOIN group_members AS gm ON gm.group_id = g.id AND gm.user_id = u.id
		WHERE g.id = $1
		
		UNION
		
		SELECT
			u.id as user_id,
			u.f_name,
			u.l_name,
			u.avatar,
			gm.role,
			gm.joined_at
		FROM group_members AS gm
		JOIN user AS u ON gm.user_id = u.id
		JOIN groups AS g ON gm.group_id = g.id
		WHERE gm.group_id = $1 AND gm.user_id != g.owner_id
		ORDER BY joined_at ASC
	`

	var members []GroupMember
	err := db.SelectContext(ctx, &members, query, groupID)
	if err != nil {
		return nil, err
	}

	return members, nil
}
