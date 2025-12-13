<div align="center">
  <img src="logo.png" alt="BrainBook Logo" width="180"/>
  <br />

# BrainBook 📖

**A Social Network for Knowledge Sharing**

[![Golang](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://golang.org/)
[![Nuxt](https://img.shields.io/badge/Nuxt-4.2-00DC82?logo=nuxt.js)](https://nuxt.com/)

</div>

---

## 📖 Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Technology Stack](#technology-stack)
- [Database Schema](#database-schema)
- [Getting Started](#getting-started)
- [API Documentation](#api-documentation)
- [Project Structure](#project-structure)
- [Authors](#authors)
- [License](#license)

---

## 🌟 Overview

BrainBook is a social networking platform to share knowledge for researchers and developers. It provides a comprehensive suite of tools for collaboration, including: user profiles, posts, groups, real-time messaging, and notifications.

### Key Highlights

- **Secure Authentication**: Session-based authentication with encrypted cookies
- **Privacy Controls**: Public and private profiles with granular post visibility
- **Real-time Communication**: WebSocket-powered instant messaging and notifications
- **Group Collaboration**: Create groups, share posts, organize events, and chat
- **Rich Media Support**: Support for images (JPEG, PNG) and GIFs in posts and comments
- **Followers System**: Private following with automatic approval for public profiles

---

## ✨ Features

### 🔐 Authentication

Complete user registration and login system with persistent sessions:

- **Registration Requirements**:

  - Email (mandatory)
  - Password (mandatory, securely hashed)
  - First Name & Last Name (mandatory)
  - Date of Birth (mandatory)
  - Avatar/Image (optional)
  - Nickname (optional)
  - About Me/Bio (optional)

- **Session Management**: Users remain logged in until they explicitly logout, with secure session tokens stored in HTTP-only cookies

### 👤 User Profiles

Each user has a comprehensive profile featuring:

- Personal information (excluding password)
- User activity and posts
- Followers and following lists
- **Profile Types**:
  - **Public Profile**: Visible to all users
  - **Private Profile**: Visible only to accepted followers
- Profile visibility toggle for authenticated users

### 👥 Followers System

Dynamic follow/unfollow functionality with request management:

- Send follow requests to other users
- Accept or decline incoming follow requests
- **Auto-follow**: Public profiles automatically accept follow requests
- View followers and following lists
- Unfollow capability for existing connections

### 📝 Posts & Comments

Create and interact with content:

- Create posts with optional images/GIFs
- Comment on posts with optional media
- **Privacy Levels**:
  - **Public**: Visible to all users
  - **Almost Private**: Visible only to accepted followers
  - **Private**: Visible only to explicitly allowed followers
- Real-time updates and interactions

### 👨‍👩‍👧‍👦 Groups

Collaborative spaces for focused discussions:

- Create groups with title and description
- Invite users to join groups
- Request to join existing groups (requires owner approval)
- Browse all available groups
- **Group Features**:
  - Group posts and comments (visible only to members)
  - Group chat rooms for real-time communication
  - Event creation and management
  - RSVP system (Going/Not Going)

### 💬 Real-time Chat

Instant messaging with WebSocket technology:

- **Private Messages**: Send messages to followers or users you're following
- **Group Chat**: Common chat room for all group members
- **Features**:
  - Real-time message delivery with offline fallback
  - Online/offline status indicators
  - Typing indicators
  - Emoji support
  - Message history

### 🔔 Notifications

Comprehensive notification system for important events:

- Follow requests (for private profiles)
- Group invitations
- Group join requests (for group owners)
- Event notifications (for group members)
- Real-time notification delivery via WebSocket

---

## 🛠️ Technology Stack

### Frontend

- **Node.js**: 22.21
- **Package Manager**: PNPM 10.24
- **Framework**: Nuxt 4.2 (Vue.js v3)
- **UI Library**: Nuxt UI Pro
- **Styling**: TailwindCSS (via Nuxt UI)
- **Utilities**: VueUse, date-fns

### Backend

- **Language**: Go 1.25
- **Database**: SQLite 3
- **Migrations**: golang-migrate/migrate
- **Security**:
  - golang.org/x/crypto/bcrypt (password hashing)
  - google/uuid (session tokens)
- **Real-time**: gorilla/websocket

### Allowed Go Dependencies

As per project requirements, only the following external Go packages are used:

- `github.com/gorilla/websocket` - WebSocket implementation
- `golang.org/x/crypto/bcrypt` - Password hashing
- `github.com/golang-migrate/migrate` - Database migrations
- `github.com/mattn/go-sqlite3` - SQLite driver
- `github.com/google/uuid` - UUID generation

---

## 🚀 Getting Started

### Prerequisites

Ensure you have the following installed:

- Go 1.25 or higher
- Node.js 22.20 or higher
- PNPM 10.18 or higher

### Installation

1. **Clone the repository**

   ```bash
   git clone https://github.com/yourusername/social-network.git
   cd social-network
   ```

2. **Backend Setup**

   ```bash
   cd backend

   # Install dependencies
   go mod download

   # Start the backend server
   go run .
   ```

   The backend server will start on `http://localhost:8080`

3. **Frontend Setup**

   ```bash
   cd frontend

   # Install dependencies
   pnpm install

   # Start the development server
   pnpm dev
   ```

   The frontend development server will start on `http://localhost:3000`

### Environment Variables

#### Backend

Create a `.env` file in the `backend` directory (optional, defaults are provided):

```env
BASE_URL=http://localhost:8080
HTTP_PORT=8080
DB_DSN=db.sqlite
DB_AUTOMIGRATE=true
```

#### Frontend

Configure API endpoint in `nuxt.config.ts` if needed.

### Building for Production

**Backend:**

```bash
cd backend
go build -o brainbook-api
./brainbook-api
```

**Frontend:**

```bash
cd frontend
pnpm build
pnpm preview
```

### Injecting Dummy Data

To populate the database with mock data for development/testing, use the provided `dummy.sql` file:

1. Make sure your SQLite database (`db.sqlite`) is initialized.
2. Inject the dummy data:

```bash
sqlite3 db.sqlite < dummy.sql
```

This will insert sample users, posts, groups, comments, and relationships themed for knowledge sharing among researchers and developers. All user passwords are set to `reboot01@BH` (bcrypt-hashed).

---

## 🗄️ Database Schema

### Database Details

The backend uses SQLite 3, with schema managed by migration files in `backend/assets/migrations/`. The main migration file is `000001_initalize_schema_migrations.up.sql`, which must match the Golang code in `internal/database/`.

#### Privacy & Visibility

- User profile privacy: `is_public` boolean field
- Post privacy: `visibility` field (`public`, `private`, `limited`)

#### Main Tables

- **user**: User profiles and account information (fields: id, f_name, l_name, email, hashed_password, dob, avatar, nickname, bio, is_public)
- **session**: Active user sessions (user_id, session_token, created_at)
- **follow_request**: Follow relationships and requests (requester_id, target_id, status)
- **post**: User posts (user_id, content, file, visibility, created_at)
- **post_user_can_view**: Private post visibility control (post_id, user_id)
- **post_comment**: Comments on posts (post_id, user_id, content, file, created_at)
- **groups**: Group information (id, owner_id, title, description, created_at)
- **group_members**: Group membership (group_id, user_id, role, joined_at)
- **group_join_requests**: Requests to join groups (group_id, requester_id, target_id, status, created_at)
- **group_posts**: Posts within groups (user_id, group_id, content, file, created_at)
- **group_post_comments**: Comments on group posts (group_post_id, user_id, content, file, created_at)
- **group_messages**: Group chat messages (group_id, sender_id, content, created_at)
- **event**: Group events (group_id, user_id, title, description, time)
- **event_has_user**: Event RSVP tracking (event_id, user_id, interested)
- **conversation**: Private conversations (user1_id, user2_id, last_message_time, created_at)
- **conversation_message**: Private messages (conversation_id, sender_id, content, created_at)
- **notifications**: Notifications (user_id, type, payload, is_read, created_at)

---

## 📚 API Documentation

Complete API documentation is available in the OpenAPI 3.1 specification file: [`openapi.yaml`](openapi.yaml)

### Public Endpoints

- `POST /v1/register` — User registration
- `POST /v1/login` — User login

### Authentication

The API uses session-based authentication with HTTP-only cookies:

1. Login via `POST /v1/login`
2. Receive `session_token` cookie
3. Include cookie in subsequent requests
4. Logout via `POST /protected/v1/logout`

---

## 👨‍💻 Authors

This project was created by:

- **Ahmed Almadhoob**
- **Abdulla Alasmawi**
- **Mohamed AlAlawi**

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- Frontend template based on [Nuxt UI Pro](https://ui.nuxt.com/).
- Built with ❤️ for Reboot Coding Institute.

---

<div align="center">

**[⬆ Back to Top](#brainbook-)**

Made with 🧠 by the BrainBook Team

</div>
