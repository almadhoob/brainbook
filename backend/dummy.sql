-- Dummy data for BrainBook
-- Password for all users: "reboot01@BH" (bcrypt hashed)

-- Users
INSERT INTO user (f_name, l_name, email, hashed_password, dob, avatar, nickname, bio, is_public) VALUES
('Alice', 'Smith', 'alice@uni.edu', '$2a$12$cYXdjXCeDZAH9u8sP0AwaOu/LjugPF17.t8J7Js0Mndz7A609spjK', '1985-04-12', NULL, 'alice', 'AI researcher and educator.', 1),
('Bob', 'Johnson', 'bob@dev.com', '$2a$12$cYXdjXCeDZAH9u8sP0AwaOu/LjugPF17.t8J7Js0Mndz7A609spjK', '1990-09-23', NULL, 'bobby', 'Full-stack developer, open source enthusiast.', 1),
('Carol', 'Lee', 'carol@data.org', '$2a$12$cYXdjXCeDZAH9u8sP0AwaOu/LjugPF17.t8J7Js0Mndz7A609spjK', '1988-02-17', NULL, 'carol', 'Data scientist, loves sharing notebooks.', 0),
('David', 'Nguyen', 'david@lab.net', '$2a$12$cYXdjXCeDZAH9u8sP0AwaOu/LjugPF17.t8J7Js0Mndz7A609spjK', '1982-12-05', NULL, 'david', 'ML researcher, conference speaker.', 1),
('Eve', 'Martins', 'eve@dev.com', '$2a$12$cYXdjXCeDZAH9u8sP0AwaOu/LjugPF17.t8J7Js0Mndz7A609spjK', '1995-07-30', NULL, 'evem', 'Frontend developer, loves Vue & Nuxt.', 1);

-- Sessions
-- INSERT INTO session (user_id, session_token) VALUES
-- (1, 'sessiontoken1'),
-- (2, 'sessiontoken2'),
-- (3, 'sessiontoken3'),
-- (4, 'sessiontoken4'),
-- (5, 'sessiontoken5');

-- Posts
INSERT INTO post (user_id, content, file, visibility) VALUES
(1, 'Sharing my latest research paper on AI ethics.', NULL, 'public'),
(2, 'Check out my new open source project for web developers!', NULL, 'public'),
(3, 'Here is a Jupyter notebook for data visualization.', NULL, 'private'),
(4, 'Slides from my ML conference talk.', NULL, 'public'),
(5, 'Tips for building scalable Vue apps.', NULL, 'public');

-- Comments
INSERT INTO post_comment (post_id, user_id, content, file) VALUES
(1, 2, 'Great work, Alice! Would love to discuss further.', NULL),
(2, 5, 'Awesome project, Bob! Starred on GitHub.', NULL),
(3, 1, 'Carol, your notebook is super helpful!', NULL),
(4, 3, 'David, can you share the code?', NULL),
(5, 4, 'Thanks for the tips, Eve!', NULL);

-- Groups
INSERT INTO groups (owner_id, title, description) VALUES
(1, 'AI Researchers', 'A group for sharing AI research, papers, and events.'),
(2, 'Web Developers', 'Discuss web technologies, frameworks, and best practices.'),
(3, 'Data Science Hub', 'Collaborate on data science projects and share resources.');

-- Group Members
INSERT INTO group_members (group_id, user_id, role) VALUES
(1, 1, 'owner'),
(1, 4, 'member'),
(2, 2, 'owner'),
(2, 5, 'member'),
(3, 3, 'owner'),
(3, 1, 'member');

-- Group Posts
INSERT INTO group_posts (user_id, group_id, content, file) VALUES
(1, 1, 'Welcome to the AI Researchers group!', NULL),
(4, 1, 'Sharing a new dataset for NLP tasks.', NULL),
(2, 2, 'Let us talk about Nuxt 4 best practices.', NULL),
(5, 2, 'Frontend performance tips for Vue.', NULL),
(3, 3, 'Data Science Hub: Weekly challenge posted!', NULL);

-- Group Post Comments
INSERT INTO group_post_comments (group_post_id, user_id, content, file) VALUES
(1, 4, 'Thanks for the welcome, Alice!', NULL),
(2, 1, 'Excited to try the dataset.', NULL),
(3, 5, 'Nuxt 4 is awesome!', NULL),
(4, 2, 'Great tips, Eve!', NULL),
(5, 1, 'Looking forward to the challenge.', NULL);

-- include created_at for deterministic ordering with new column
INSERT INTO follow_request (requester_id, target_id, status, created_at) VALUES
(2, 1, 'accepted', '2025-12-01 09:00:00'),
(3, 1, 'pending',   '2025-12-02 10:00:00'),
(4, 3, 'accepted', '2025-12-03 11:00:00'),
(5, 2, 'accepted', '2025-12-04 12:00:00');

-- Conversations
INSERT INTO conversation (user1_id, user2_id) VALUES
(1, 2),
(3, 4);

-- Conversation Messages
INSERT INTO conversation_message (conversation_id, sender_id, content) VALUES
(1, 1, 'Hi Bob, let us collaborate on the AI project.'),
(1, 2, 'Sure Alice, I''m interested!'),
(2, 3, 'David, your ML talk was inspiring.'),
(2, 4, 'Thanks Carol! Happy to share resources.');

-- Group Messages
INSERT INTO group_messages (group_id, sender_id, content) VALUES
(1, 1, 'Welcome everyone!'),
(1, 4, 'Glad to be here.'),
(2, 2, 'Web devs unite!'),
(2, 5, 'Vue is the best.'),
(3, 3, 'Weekly challenge: Predict housing prices.');

-- Events
INSERT INTO event (group_id, user_id, title, description, time) VALUES
(1, 1, 'AI Ethics Webinar', 'Join us for a discussion on AI ethics.', '2025-12-10 18:00:00'),
(2, 2, 'Nuxt 4 Workshop', 'Hands-on workshop for Nuxt 4.', '2025-12-15 15:00:00'),
(3, 3, 'Data Science Challenge', 'Participate in our weekly challenge.', '2025-12-20 20:00:00');

-- Event RSVPs
INSERT INTO event_has_user (event_id, user_id, interested) VALUES
(1, 1, 1),
(1, 4, 1),
(2, 2, 1),
(2, 5, 1),
(3, 3, 1),
(3, 1, 1);

-- Notifications
INSERT INTO notifications (user_id, type, payload, is_read) VALUES
(1, 'follow_request', '{"from":2}', 0),
(1, 'comment', '{"post_id":1,"from":2}', 0),
(2, 'group_invite', '{"group_id":2,"from":5}', 0),
(3, 'event', '{"event_id":3}', 0),
(4, 'message', '{"conversation_id":2,"from":3}', 0),
(5, 'comment', '{"post_id":2,"from":5}', 0);
