package models

import (
	"net/http"
	"time"
)

// ZitadelClient provides a wrapper around the Zitadel Management API.
type ZitadelClient struct {
	BaseURL      string
	ServiceToken string
	ProjectID    string
	HTTPClient   *http.Client
}

// ZitadelConfig holds the configuration for the Zitadel API client.
type ZitadelConfig struct {
	IssuerURL    string
	ServiceToken string
	ProjectID    string
	Timeout      time.Duration
}

// CreateUserRequest contains the parameters for creating a new user.
type CreateUserRequest struct {
	Email string `json:"email"`
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	Password string `json:"password,omitempty"`
	OrgID string `json:"org_id"`
}

// CreateUserResponse contains the response from creating a user.
type CreateUserResponse struct {
	UserID string `json:"user_id"`
	Email string `json:"email"`
}

// UserResponse contains the details of a user retrieved from Zitadel.
type UserResponse struct {
	UserID string `json:"user_id"`
	UserName string `json:"user_name"`
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	Email string `json:"email"`
	State string `json:"state"`
}

// ZitadelCreateUserRequestBody is the internal API request format.
type ZitadelCreateUserRequestBody struct {
	UserName string `json:"username"`
	Profile  struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	} `json:"profile"`
	Email struct {
		Email      string `json:"email"`
		IsVerified bool   `json:"is_email_verified"`
	} `json:"email"`
	Password string `json:"password,omitempty"`
}

// ZitadelCreateUserResponseBody is the internal API response format.
type ZitadelCreateUserResponseBody struct {
	UserID  string `json:"user_id"`
	Details struct {
		Sequence      string    `json:"sequence"`
		CreationDate  time.Time `json:"creation_date"`
		ChangeDate    time.Time `json:"change_date"`
		ResourceOwner string    `json:"resource_owner"`
	} `json:"details"`
}

// ZitadelRoleAssignment is the internal API format for role assignments.
type ZitadelRoleAssignment struct {
	RoleKeys []string `json:"role_keys"`
}

// ZitadelGetUserResponseBody is the internal API response format for GetUser.
type ZitadelGetUserResponseBody struct {
	User struct {
		UserID   string `json:"user_id"`
		UserName string `json:"user_name"`
		State    string `json:"state"`
		Human    struct {
			Profile struct {
				FirstName string `json:"first_name"`
				LastName  string `json:"last_name"`
			} `json:"profile"`
			Email struct {
				Email string `json:"email"`
			} `json:"email"`
		} `json:"human"`
	} `json:"user"`
}
