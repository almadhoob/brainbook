package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"brainbook-api/internal/cookie"
)

type User struct {
	ID             int       `db:"id" json:"id"`
	FName          string    `db:"f_name" json:"f_name"`
	LName          string    `db:"l_name" json:"l_name"`
	Email          string    `db:"email" json:"email"`
	HashedPassword string    `db:"hashed_password" json:"-"`
	DOB            time.Time `db:"dob" json:"dob"`
	Avatar         []byte    `db:"avatar" json:"avatar"`
	Nickname       string    `db:"nickname" json:"nickname"`
	Bio            string    `db:"bio" json:"bio"`
	IsPublic       bool      `db:"is_public" json:"is_public"`
}

func (u *User) FullName() string {
	return u.FName + " " + u.LName
}

type UserSummary struct {
	ID     int    `db:"user_id" json:"user_id"`
	FName  string `db:"f_name" json:"f_name"`
	LName  string `db:"l_name" json:"l_name"`
	Avatar []byte `db:"avatar" json:"avatar"`
}

func (u *UserSummary) FullName() string {
	return u.FName + " " + u.LName
}

// UserWithLastMessage represents a user with their last message time
type UserWithLastMessageTime struct {
	LastMessageTime *string `db:"last_message_time" json:"last_message_time"`

	UserSummary
}

type UserPatch struct {
	Avatar   *[]byte `json:"avatar"`
	Nickname *string `json:"nickname"`
	Bio      *string `json:"bio"`
}

// Checks if the context user and target user are the same.
func (user *User) IsUserIDMatching(targetUserID int) bool {
	return user.ID == targetUserID
}



func (db *DB) InsertUser(firstName, lastName, email, hashedPassword, nickname, bio string, dob time.Time, avatar []byte) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
    INSERT INTO user (f_name, l_name, email, hashed_password, dob, avatar, nickname, bio)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	result, err := db.ExecContext(ctx, query, firstName, lastName, email, hashedPassword, dob, avatar, nickname, bio)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), err
}

func (db *DB) UserById(id int) (*User, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var user User

	query := `SELECT * FROM user WHERE id = $1`

	err := db.GetContext(ctx, &user, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}

	return &user, true, err
}

func (db *DB) UserByEmail(email string) (*User, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var user User

	query := `SELECT * FROM user WHERE email = $1`

	err := db.GetContext(ctx, &user, query, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}

	return &user, true, err
}

func (db *DB) UserByUsername(username string) (*User, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var user User

	query := `SELECT * FROM user WHERE username = $1`

	err := db.GetContext(ctx, &user, query, username)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}

	return &user, true, err
}

func (db *DB) UserBySession(sessionToken string) (*User, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var userID int

	// Conversion to int64 is necessary for SQL compatibility
	expiryMinutes := int(cookie.CookieExpirey.Minutes())
	query := fmt.Sprintf(`
    SELECT user_id 
    FROM session 
    WHERE session_token = $1 
    AND datetime(created_at, '+%d minutes') > datetime('now')`, expiryMinutes)

	err := db.GetContext(ctx, &userID, query, sessionToken)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	// Use existing Getuser.User function
	return db.UserById(userID)
}

func (db *DB) UpdateUserHashedPassword(userID int, hashedPassword string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `UPDATE user SET hashed_password = $1 WHERE id = $2`

	_, err := db.ExecContext(ctx, query, hashedPassword, userID)
	return err
}

func (db *DB) UpdateBio(userID int, bio string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `UPDATE user SET bio = $1 WHERE id = $2`

	_, err := db.ExecContext(ctx, query, bio, userID)
	return err
}

func (db *DB) UpdateNickname(userID int, nickname string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `UPDATE user SET nickname = $1 WHERE id = $2`

	_, err := db.ExecContext(ctx, query, nickname, userID)
	return err
}

func (db *DB) UpdateAvatar(userID int, avatar []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `UPDATE user SET avatar = $1 WHERE id = $2`

	_, err := db.ExecContext(ctx, query, avatar, userID)
	return err
}

func (db *DB) UpdatePrivacy(userID int, isPrivate bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `UPDATE user SET is_private = $1 WHERE id = $2`

	_, err := db.ExecContext(ctx, query, isPrivate, userID)
	return err
}

func (db *DB) PendingFollowRequestsCount(userID int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var count int
	query := `SELECT COUNT(*) FROM follow_request WHERE target_id = $1 AND status = 'pending'`

	err := db.GetContext(ctx, &count, query, userID)
	if err != nil {
		return 0, err
	}

	return count, nil
}


// GetTotaluser.UserCountExcludinguser.User returns the total number of users excluding a specific user
func (db *DB) TotalUserCount() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var count int
	query := `SELECT COUNT(*) FROM user`

	err := db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, err
	}

	return count - 1, nil
}

// TO DO: Instead of this atrocious function, the conversation table can now be used
// to retrieve rcently messaged users first. Any other users not paired with the context user in the
// conversation table can then be display independently in whataever order.

// UsersList returns users ordered by recent chat activity, then alphabetically
func (db *DB) UserList(currentUserID int) ([]UserWithLastMessageTime, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		SELECT u.id, u.f_name, u.l_name 
			CASE WHEN MAX(m.created_at) IS NOT NULL 
				THEN datetime(MAX(m.created_at))
				ELSE NULL 
			END as last_message_time
		FROM user u
		LEFT JOIN message m ON u.id = (m.sender_id
			OR m.sender_id = $1)
		WHERE u.id != $1
		GROUP BY u.id, u.username
		ORDER BY 
			CASE WHEN MAX(m.created_at) IS NOT NULL THEN 0 ELSE 1 END,
			MAX(m.created_at) DESC,
			u.username ASC`

	var users []UserWithLastMessageTime
	err := db.SelectContext(ctx, &users, query, currentUserID)
	if err != nil {
		return nil, err
	}

	return users, nil
}


func (db *DB) FollowerCountByUserID(userID int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var count int
	query := `SELECT COUNT(*) FROM follow_request WHERE target_id = $1 AND status = 'accepted'`

	err := db.GetContext(ctx, &count, query, userID)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (db *DB) FollowingCountByUserID(userID int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var count int
	query := `SELECT COUNT(*) FROM follow_request WHERE requester_id = $1 AND status = 'accepted'`

	err := db.GetContext(ctx, &count, query, userID)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (db *DB) IsFollowing(requesterID, targetID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var count int
	query := `SELECT COUNT(*) FROM follow_request WHERE requester_id = $1 AND target_id = $2 AND status = 'accepted'`

	err := db.GetContext(ctx, &count, query, requesterID, targetID)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (db *DB) FollowersByUserID(userID int) ([]UserSummary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var followers []UserSummary
	query := `
		SELECT u.id as user_id, u.f_name, u.l_name, u.avatar
		FROM user u
		JOIN follow_request fr ON u.id = fr.requester_id
		WHERE fr.target_id = $1 AND fr.status = 'accepted'`

	err := db.SelectContext(ctx, &followers, query, userID)
	if err != nil {
		return nil, err
	}

	return followers, nil
}

func (db *DB) FollowingByUserID(userID int) ([]UserSummary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var following []UserSummary
	query := `
		SELECT u.id as user_id, u.f_name, u.l_name, u.avatar
		FROM user u
		JOIN follow_request fr ON u.id = fr.target_id
		WHERE fr.requester_id = $1 AND fr.status = 'accepted'`

	err := db.SelectContext(ctx, &following, query, userID)
	if err != nil {
		return nil, err
	}

	return following, nil
}