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

	// ErrMessageNotFound indicates message does not exist
	ErrMessageNotFound = errors.New("message not found")

	// ErrInvalidMessageText indicates message text is invalid (empty or too long)
	ErrInvalidMessageText = errors.New("message text must be between 1 and 10,000 characters")

	// ErrInviteNotFound indicates invite token does not exist or is invalid
	ErrInviteNotFound = errors.New("invite not found")

	// ErrInviteExpired indicates invite token has expired
	ErrInviteExpired = errors.New("invite has expired")

	// ErrInviteAlreadyConsumed indicates invite token has already been used
	ErrInviteAlreadyConsumed = errors.New("invite has already been consumed")

	// ErrInstanceAlreadyBootstrapped indicates bootstrap has already been completed
	ErrInstanceAlreadyBootstrapped = errors.New("instance has already been bootstrapped")

	// ErrInsufficientPermissions indicates user lacks required permissions
	ErrInsufficientPermissions = errors.New("insufficient permissions")

	// ErrInvalidDMTarget indicates attempting to create DM with invalid user
	ErrInvalidDMTarget = errors.New("cannot create DM with yourself")

	// ErrUnauthorizedMessageEdit indicates user cannot edit this message
	ErrUnauthorizedMessageEdit = errors.New("can only edit your own messages")

	// ErrUnauthorizedMessageDelete indicates user cannot delete this message
	ErrUnauthorizedMessageDelete = errors.New("can only delete your own messages")

	// ErrEntityLimitReached indicates workspace entity cap is reached for current plan.
	ErrEntityLimitReached = errors.New("entity limit reached")

	// ErrUnauthorized indicates authentication/authorization failed (e.g. invalid central JWT)
	ErrUnauthorized = errors.New("unauthorized")
)
