package authkit

import "errors"

// Invitation and access code errors.
var (
	// ErrInvalidAccessCode is returned when an access code format is invalid.
	ErrInvalidAccessCode = errors.New("invalid access code format")

	// ErrAccessCodeExpired is returned when an access code has expired.
	ErrAccessCodeExpired = errors.New("access code has expired")

	// ErrAccessCodeUsed is returned when an access code has already been used.
	ErrAccessCodeUsed = errors.New("access code has already been used")

	// ErrInvitationNotFound is returned when an invitation cannot be found.
	ErrInvitationNotFound = errors.New("invitation not found")

	// ErrInvitationExpired is returned when an invitation has expired.
	ErrInvitationExpired = errors.New("invitation has expired")
)

// Organization errors.
var (
	// ErrUnauthorizedOrgAction is returned when a user attempts an action
	// they don't have permission for within an organization.
	ErrUnauthorizedOrgAction = errors.New("unauthorized organization action")

	// ErrUserAlreadyExists is returned when attempting to create a user
	// that already exists in the identity provider.
	ErrUserAlreadyExists = errors.New("user already exists")
)
