# Project Requirements

## Objectives

You have to create a social network that will contain the following features:
- Authentication
- Profiles
- Followers
- Posts
- Groups
- Chats
- Notifications

## Instructions

### Frontend

You have to use VueJS v3 (with Nuxt v4). Responsiveness and performance are main objectives of the frontend.

### Backend

#### API

The backend consists of an API in Golang. This logic will have several middleware including:
- Authentication, using sessions and cookies.
- Images handling, with JPEG, PNG and GIF types.
- Websocket, for connections in real time between clients.

For the web server you, you can use Caddy Server or use your own web server.

You are only allowed to use Golang standard library in addition to the following:
- gorilla/websocket
- crypto/bcrypt
- golang-migrate
- Boostport/migration
- rubenv/sql-migrate
- mattn/go-sqlite3
- google/uuid
- gofrs/uuid

#### Database

You will use SQLite. You have to create database migrations using golang-migrate package.

## Features

### Authentication

In order for the users to use the social network they have to make an account. So you have to make a registration and login form. To register, every user must provide:
- Email (Mandatory)
- Password (Mandatory)
- First Name (Mandatory)
- Last Name (Mandatory)
- Date of Birth (Mandatory)
- Avatar/Image (Optional)
- Nickname (Optional)
- About Me (Optional)

Note that the **Avatar/Image**, **Nickname** and **About Me** must be present in the form but the user can skip the filling of those fields.

When the user logins, he must stays logged in until he chooses a logout option that must be available at all times. For this you have to implement sessions and cookies.

### Profile

Every profile must contain:
- User information (every information requested in the register form except **Password**)
- User activity
  * Every post made by the user
- Followers and following users (display the users that are following the owner of the profile and who he is following)

There are two types of profiles: a public profile and a private profile. A public profile will display the information specified above to every user on the social network, while the private profile will only display that same information to their followers only.

When the user is in their own profile it must be available an option that allows the user to turn its profile public or private.

### Followers

When navigating the social network, the user must be able to follow and unfollow other users. Needless to say that to unfollow a user you have to be following him.

Regarding following someone, the user must initiate this by sending a follow request to the desired user. The recipient user can then choose to "accept" or "decline" the request. However, if the recipient user has a public profile, this request-and-accept process is bypassed and the user who sent the request automatically starts following the user with the pubic profile.

### Posts

After a user is logged in, he can create posts and comments on already created posts. While creating a post or a comment, the user can include an image (including GIF).

The user must be able to specify the privacy of the post:
- public (all users in the social network will be able to see the post)
- almost private (only followers of the creator of the post will be able to see the post)
- private (only the followers chosen by the creator of the post will be able to see it)

### Groups

A user must be able to create a group. The group must have a title and a description given by the creator, and he can invite other users to join the group.

The invited users need to accept the invitation to be part of the group. They can also invite other people once they are already part of the group. Another way to enter the group is to request to be in it and only the creator of the group would be allowed to accept or refuse the request.

To make a request to enter a group the user must find it first. This will be possible by having a section where you can browse through all groups.

When in a group, a user can create posts and comment the posts already created. These posts and comments will only be displayed to members of the group.

A user belonging to the group can also create an event, making it available for the other group users. An event must have:
- Title
- Description
- Day/Time
- RSVP Options:
  * Going
  * Not going

After creating the event every user can choose one of the RSVP options for the event.

### Chat

Users must be able to send private messages to other users that they are following or being followed, in other words, at least one of the users must be following the other.

When a user sends a message, the recipient will instantly receive it through Websockets if they are following the sender or if the recipient has a public profile.

It must be able for the users to send emojis to each other.

Groups must have a common chat room, so if a user is a member of the group he must be able to send and receive messages to this group chat.

### Notifications

A user must be able to see the notifications in every page of the project. New notifications are different from new private messages and must be displayed in a different way.

A user must be notified if he:
- has a private profile and some other user sends him/her a following request
- receives a group invitation, so he can refuse or accept the request
- is the creator of a group and another user requests to join the group, so he can refuse or accept the request
- is member of a group and an event is created
