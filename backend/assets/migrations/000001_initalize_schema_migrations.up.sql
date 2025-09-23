CREATE TABLE IF NOT EXISTS user (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    f_name TEXT NOT NULL,
    l_name TEXT NOT NULL,
    email TEXT NOT NULL,
    password TEXT NOT NULL,
    dob DATETIME NOT NULL,
    avatar BLOB,
    nickname TEXT,
    bio TEXT,
    is_public BOOLEAN NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS post (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    content TEXT,
    file BLOB,
    visibility CHECK( visibility IN ('public','private','limited') ) NOT NULL DEFAULT 'public',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user(id)
);

CREATE TABLE IF NOT EXISTS post_user_can_view (
    post_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    PRIMARY KEY (post_id, user_id),
    FOREIGN KEY (post_id) REFERENCES post(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS post_comment (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    content TEXT,
    file BLOB,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (post_id) REFERENCES post(id),
    FOREIGN KEY (user_id) REFERENCES user(id)
);

CREATE TABLE IF NOT EXISTS follow_request (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    requester_id INTEGER NOT NULL,
    target_id INTEGER NOT NULL,
    status CHECK( status IN ('pending','accepted','declined') ) NOT NULL DEFAULT 'pending',
    FOREIGN KEY (requester_id) REFERENCES user(id),
    FOREIGN KEY (target_id) REFERENCES user(id)
);

CREATE TABLE IF NOT EXISTS session (
    user_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    session_token TEXT PRIMARY KEY,
    FOREIGN KEY (user_id) REFERENCES user(id)
);

CREATE TABLE IF NOT EXISTS group (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_id INTEGER NOT NULL,
    title TEXT,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_id) REFERENCES user(id)
);

CREATE TABLE IF NOT EXISTS group_member (
    group_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    role TEXT CHECK( role IN ('member','owner') ) NOT NULL DEFAULT 'member',
    joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (group_id, user_id),
    FOREIGN KEY (group_id) REFERENCES "group"(id),
    FOREIGN KEY (user_id) REFERENCES user(id)
);

CREATE TABLE IF NOT EXISTS group_join_request (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL,
    requester_id INTEGER NOT NULL,
    status CHECK( status IN ('pending','accepted','declined') ) NOT NULL DEFAULT 'pending',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES "group"(id),
    FOREIGN KEY (requester_id) REFERENCES user(id)
);

CREATE TABLE IF NOT EXISTS group_post (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    group_id INTEGER NOT NULL,
    content TEXT,
    file BLOB,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user(id),
    FOREIGN KEY (group_id) REFERENCES "group"(id)
);

CREATE TABLE IF NOT EXISTS group_post_comment (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_post_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    content TEXT,
    file BLOB,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_post_id) REFERENCES group_post(id),
    FOREIGN KEY (user_id) REFERENCES user(id)
);

CREATE TABLE IF NOT EXISTS private_conversation (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user1_id INTEGER NOT NULL,
    user2_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user1_id) REFERENCES user(id),
    FOREIGN KEY (user2_id) REFERENCES user(id)
);

CREATE TABLE IF NOT EXISTS private_message (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id INTEGER NOT NULL,
    sender_id INTEGER NOT NULL,
    content TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (conversation_id) REFERENCES private_conversation(id),
    FOREIGN KEY (sender_id) REFERENCES user(id)
);

CREATE TABLE IF NOT EXISTS group_message (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL,
    sender_id INTEGER NOT NULL,
    content TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES "group"(id),
    FOREIGN KEY (sender_id) REFERENCES user(id)
);

CREATE TABLE IF NOT EXISTS event (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    time DATETIME NOT NULL,
    FOREIGN KEY (group_id) REFERENCES "group"(id)
);

CREATE TABLE IF NOT EXISTS event_has_user (
    event_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    interested BOOLEAN NOT NULL,
    PRIMARY KEY (event_id, user_id),
    FOREIGN KEY (event_id) REFERENCES event(id),
    FOREIGN KEY (user_id) REFERENCES user(id)
);
