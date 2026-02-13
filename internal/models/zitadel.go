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
	UserName string `json:"username"`
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
	UserID    string `json:"user_id"`
	UserName  string `json:"user_name"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	State     string `json:"state"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// ZitadelCreateUserRequestBody is the internal API request format.
type ZitadelCreateUserRequestBody struct {
	UserName string `json:"userName"`
	Profile  struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	} `json:"profile"`
	Email struct {
		Email      string `json:"email"`
		IsVerified bool   `json:"isEmailVerified"`
	} `json:"email"`
	Password string `json:"password,omitempty"`
}

// ZitadelCreateUserResponseBody is the internal API response format.
type ZitadelCreateUserResponseBody struct {
	UserID  string `json:"userId"`
	Details struct {
		Sequence      string    `json:"sequence"`
		CreationDate  time.Time `json:"creationDate"`
		ChangeDate    time.Time `json:"changeDate"`
		ResourceOwner string    `json:"resourceOwner"`
	} `json:"details"`
}

// ZitadelRoleAssignment is the internal API format for role assignments.
type ZitadelRoleAssignment struct {
	ProjectID string   `json:"projectId"`
	RoleKeys  []string `json:"roleKeys"`
}

// ZitadelGetUserResponseBody is the internal API response format for GetUser.
type ZitadelGetUserResponseBody struct {
	User struct {
		UserID   string `json:"userId"`
		UserName string `json:"userName"`
		State    string `json:"state"`
		Human    struct {
			Profile struct {
				FirstName   string `json:"firstName"`
				LastName    string `json:"lastName"`
				DisplayName string `json:"displayName"`
				AvatarURL   string `json:"avatarUrl"`
			} `json:"profile"`
			Email struct {
				Email string `json:"email"`
			} `json:"email"`
		} `json:"human"`
	} `json:"user"`
}

// OrganizationResponse contains the details of an organization retrieved from Zitadel.
type OrganizationResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	State       string `json:"state"`
	Domain      string `json:"domain,omitempty"`
	LogoURL     string `json:"logo_url,omitempty"`
	LogoDarkURL string `json:"logo_dark_url,omitempty"`
	IconURL     string `json:"icon_url,omitempty"`
	IconDarkURL string `json:"icon_dark_url,omitempty"`
}

// ZitadelLabelPolicyResponseBody is the response format for getting org label/branding policy
type ZitadelLabelPolicyResponseBody struct {
	Policy struct {
		LogoURL         string `json:"logoUrl"`
		LogoDarkURL     string `json:"logoUrlDark"`
		IconURL         string `json:"iconUrl"`
		IconDarkURL     string `json:"iconUrlDark"`
		PrimaryColor    string `json:"primaryColor"`
		BackgroundColor string `json:"backgroundColor"`
		FontURL         string `json:"fontUrl"`
	} `json:"policy"`
}

// ZitadelGetOrgResponseBody is the internal API response format for GetOrganization.
type ZitadelGetOrgResponseBody struct {
	Org struct {
		ID            string `json:"id"`
		Name          string `json:"name"`
		State         string `json:"state"`
		PrimaryDomain string `json:"primaryDomain"`
	} `json:"org"`
}

// UserGrant represents a user's grant (role assignment) in a project
type UserGrant struct {
	ID        string   `json:"id"`
	UserID    string   `json:"user_id"`
	ProjectID string   `json:"project_id"`
	RoleKeys  []string `json:"role_keys"`
	State     string   `json:"state"`
	// User details (populated from separate call or included in response)
	UserName  string `json:"user_name,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Email     string `json:"email,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// ZitadelListUserGrantsResponseBody is the response format for listing user grants
type ZitadelListUserGrantsResponseBody struct {
	Result []struct {
		ID           string   `json:"id"`
		UserID       string   `json:"userId"`
		ProjectID    string   `json:"projectId"`
		RoleKeys     []string `json:"roleKeys"`
		State        string   `json:"state"`
		UserName     string   `json:"userName"`
		FirstName    string   `json:"firstName"`
		LastName     string   `json:"lastName"`
		Email        string   `json:"email"`
		AvatarURL    string   `json:"avatarUrl"`
		DisplayName  string   `json:"displayName"`
		OrgID        string   `json:"orgId"`
		ProjectName  string   `json:"projectName"`
		GrantedOrgID string   `json:"grantedOrgId"`
	} `json:"result"`
	Details struct {
		TotalResult string `json:"totalResult"`
	} `json:"details"`
}

// OrgMetadata represents organization metadata (for logo, location, etc.)
type OrgMetadata struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ZitadelListOrgMetadataResponseBody is the response format for listing org metadata
type ZitadelListOrgMetadataResponseBody struct {
	Result []struct {
		Key   string `json:"key"`
		Value string `json:"value"` // Base64 encoded
	} `json:"result"`
}

// ZitadelSearchUsersRequestBody is the request format for searching users
type ZitadelSearchUsersRequestBody struct {
	Query struct {
		Offset uint64 `json:"offset,omitempty"`
		Limit  uint32 `json:"limit,omitempty"`
		Asc    bool   `json:"asc,omitempty"`
	} `json:"query,omitempty"`
	Queries []interface{} `json:"queries,omitempty"`
}

// ZitadelSearchUsersResponseBody is the response format for searching users
type ZitadelSearchUsersResponseBody struct {
	Result []struct {
		UserID   string `json:"userId"`
		UserName string `json:"userName"`
		State    string `json:"state"`
		Human    *struct {
			Profile struct {
				FirstName   string `json:"firstName"`
				LastName    string `json:"lastName"`
				DisplayName string `json:"displayName"`
			} `json:"profile"`
			Email struct {
				Email      string `json:"email"`
				IsVerified bool   `json:"isEmailVerified"`
			} `json:"email"`
		} `json:"human,omitempty"`
	} `json:"result"`
	Details struct {
		TotalResult string `json:"totalResult"`
	} `json:"details"`
}

// IDPLink represents a user's linked external identity provider
type IDPLink struct {
	IDPID          string `json:"idp_id"`
	IDPName        string `json:"idp_name"`
	UserID         string `json:"user_id"`
	ExternalUserID string `json:"external_user_id"`
	DisplayName    string `json:"display_name"`
	ProvidedUserID string `json:"provided_user_id,omitempty"`
	ProvidedEmail  string `json:"provided_email,omitempty"`
}

// ZitadelListIDPLinksResponseBody is the response format for listing user IDP links
type ZitadelListIDPLinksResponseBody struct {
	Result []struct {
		IDPID          string `json:"idpId"`
		UserID         string `json:"userId"`
		IDPName        string `json:"idpName"`
		ProvidedUserID string `json:"providedUserId"`
		ProvidedEmail  string `json:"providedUserName"` // Zitadel uses providedUserName for email
		IDPType        int    `json:"idpType"`
	} `json:"result"`
	Details struct {
		TotalResult string `json:"totalResult"`
	} `json:"details"`
}
