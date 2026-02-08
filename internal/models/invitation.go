package models

import "time"

// InvitationClaims represents the claims embedded in an invitation token.
type InvitationClaims struct {
	InvitationID string    `json:"invitation_id"`
	OrgID        string    `json:"org_id"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedBy    string    `json:"created_by"`
}

// AccessCode represents a secure one-time access code for invitation acceptance.
type AccessCode struct {
	Code         string     `json:"code"`
	InvitationID string     `json:"invitation_id"`
	ExpiresAt    time.Time  `json:"expires_at"`
	UsedAt       *time.Time `json:"used_at,omitempty"`
}
