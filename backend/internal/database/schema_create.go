package database

import (
	_ "github.com/mattn/go-sqlite3"
)

func (db DB) CreateDatabase() error {

	const CreateUserTable = `
    CREATE TABLE IF NOT EXISTS user (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        email TEXT NOT NULL,
        password TEXT NOT NULL,
        f_name TEXT NOT NULL,
        l_name TEXT NOT NULL,
        date_of_birth DATETIME NOT NULL,
        avatar BLOB,
        nickname TEXT,
        bio TEXT,
        is_public BOOLEAN NOT NULL DEFAULT 1
    );`

	const CreatePostTable = `
    CREATE TABLE IF NOT EXISTS post (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        content TEXT,
        image BLOB,
        privacy INTEGER, 
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES user(id)
    );`

	const CreatePostCommentTable = `
    CREATE TABLE IF NOT EXISTS post_comment (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        content TEXT,
        image BLOB,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        post_id INTEGER NOT NULL,
        user_id INTEGER NOT NULL,
        FOREIGN KEY (post_id) REFERENCES post(id),
        FOREIGN KEY (user_id) REFERENCES user(id)
    );`

	const CreateFollowRequestTable = `
    CREATE TABLE IF NOT EXISTS follow_request (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        requester_id INTEGER NOT NULL,
        target_id INTEGER NOT NULL,
        status INTEGER NOT NULL,
        FOREIGN KEY (requester_id) REFERENCES user(id),
        FOREIGN KEY (target_id) REFERENCES user(id)
    );`

	const CreateSessionTable = `
    CREATE TABLE IF NOT EXISTS session (
        session_token TEXT PRIMARY KEY,
        user_id INTEGER NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES user(id)
    );`

	const CreateGroupTable = `
    CREATE TABLE IF NOT EXISTS group_table (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        creator_id INTEGER NOT NULL,
        title TEXT,
        description TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (creator_id) REFERENCES user(id)
    );`

	const CreateGroupMemberTable = `
    CREATE TABLE IF NOT EXISTS group_member (
        group_id INTEGER NOT NULL,
        user_id INTEGER NOT NULL,
        role INTEGER NOT NULL, 
        joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        PRIMARY KEY (group_id, user_id),
        FOREIGN KEY (group_id) REFERENCES group_table(id),
        FOREIGN KEY (user_id) REFERENCES user(id)
    );`

	const CreateGroupJoinRequestTable = `
    CREATE TABLE IF NOT EXISTS group_join_request (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        group_id INTEGER NOT NULL,
        requester_id INTEGER NOT NULL,
        status INTEGER NOT NULL, -- 0 = pending, 1 = accepted, 2 = declined
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (group_id) REFERENCES group_table(id),
        FOREIGN KEY (requester_id) REFERENCES user(id)
    );`

	const CreateGroupPostTable = `
    CREATE TABLE IF NOT EXISTS group_post (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        content TEXT,
        image BLOB,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        user_id INTEGER NOT NULL,
        group_id INTEGER NOT NULL,
        FOREIGN KEY (user_id) REFERENCES user(id),
        FOREIGN KEY (group_id) REFERENCES group_table(id)
    );`

	const CreateGroupCommentTable = `
    CREATE TABLE IF NOT EXISTS group_comment (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        content TEXT,
        image BLOB,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        group_post_id INTEGER NOT NULL,
        user_id INTEGER NOT NULL,
        FOREIGN KEY (group_post_id) REFERENCES group_post(id),
        FOREIGN KEY (user_id) REFERENCES user(id)
    );`

	const CreatePrivateConversationTable = `
    CREATE TABLE IF NOT EXISTS private_conversation (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user1_id INTEGER NOT NULL,
        user2_id INTEGER NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (user1_id) REFERENCES user(id),
        FOREIGN KEY (user2_id) REFERENCES user(id)
    );`

	const CreatePrivateMessageTable = `
    CREATE TABLE IF NOT EXISTS private_message (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        conversation_id INTEGER NOT NULL,
        sender_id INTEGER NOT NULL,
        content TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (conversation_id) REFERENCES private_conversation(id),
        FOREIGN KEY (sender_id) REFERENCES user(id)
    );`

	const CreateGroupMessageTable = `
    CREATE TABLE IF NOT EXISTS group_message (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        group_id INTEGER NOT NULL,
        sender_id INTEGER NOT NULL,
        content TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (group_id) REFERENCES group_table(id),
        FOREIGN KEY (sender_id) REFERENCES user(id)
    );`

	const CreateEventTable = `
    CREATE TABLE IF NOT EXISTS event (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
        description TEXT,
        time DATETIME NOT NULL,
        group_id INTEGER NOT NULL,
        FOREIGN KEY (group_id) REFERENCES group_table(id)
    );`

	const CreateEventHasUserTable = `
    CREATE TABLE IF NOT EXISTS event_has_user (
        event_id INTEGER NOT NULL,
        user_id INTEGER NOT NULL,
        interested BOOLEAN NOT NULL, -- true = going, false = not going
        PRIMARY KEY (event_id, user_id),
        FOREIGN KEY (event_id) REFERENCES event(id),
        FOREIGN KEY (user_id) REFERENCES user(id)
    );`

	createTableStatements := []string{
		CreateUserTable,
		CreatePostTable,
		CreatePostCommentTable,
		CreateFollowRequestTable,
		CreateSessionTable,
		CreateGroupTable,
		CreateGroupMemberTable,
		CreateGroupJoinRequestTable,
		CreateGroupPostTable,
		CreateGroupCommentTable,
		CreatePrivateConversationTable,
		CreatePrivateMessageTable,
		CreateGroupMessageTable,
		CreateEventTable,
		CreateEventHasUserTable,
	}

	for _, stmt := range createTableStatements {
		_, err := db.Exec(stmt)
		if err != nil {
			return err
		}
	}

	return nil
}
