package models

import "errors"

// Authentication errors
var (
	// ErrUserNotFound indicates user does not exist
	ErrUserNotFound = errors.New("user not found")

	// ErrEmailTaken indicates email is already registered
	ErrEmailTaken = errors.New("email already taken")

	// ErrInvalidCredentials indicates authentication failed
	// (intentionally vague to not leak whether email or password is wrong)
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrInvalidEmail indicates email format is invalid
	ErrInvalidEmail = errors.New("invalid email format")

	// ErrWeakPassword indicates password does not meet requirements
	ErrWeakPassword = errors.New("password must be at least 8 characters with 1 letter and 1 number")

	// ErrInvalidToken indicates JWT token is invalid or expired
	ErrInvalidToken = errors.New("invalid or expired token")

	// ErrChannelNotFound indicates channel does not exist
	ErrChannelNotFound = errors.New("channel not found")

	// ErrChannelNameTaken indicates channel name is already in use
	ErrChannelNameTaken = errors.New("channel name already taken")

	// ErrInvalidChannelName indicates channel name format is invalid
	ErrInvalidChannelName = errors.New("channel name must be lowercase alphanumeric with hyphens only")

	// ErrNotChannelMember indicates user is not a member of the channel
	ErrNotChannelMember = errors.New("not a member of this channel")

	// ErrAlreadyChannelMember indicates user is already a member
	ErrAlreadyChannelMember = errors.New("already a member of this channel")
)
