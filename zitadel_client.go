package authkit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Prescott-Data/dromos-authkit/internal/models"
)

// ZitadelClient wraps models.ZitadelClient for convenience.
type ZitadelClient struct {
	*models.ZitadelClient
}

// ZitadelConfig is an alias to models.ZitadelConfig for backward compatibility.
type ZitadelConfig = models.ZitadelConfig

// CreateUserRequest is an alias to models.CreateUserRequest for backward compatibility.
type CreateUserRequest = models.CreateUserRequest

// CreateUserResponse is an alias to models.CreateUserResponse for backward compatibility.
type CreateUserResponse = models.CreateUserResponse

// UserResponse is an alias to models.UserResponse for backward compatibility.
type UserResponse = models.UserResponse

// NewZitadelClient creates a new Zitadel API client with the provided configuration.
func NewZitadelClient(cfg ZitadelConfig) *ZitadelClient {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &ZitadelClient{
		ZitadelClient: &models.ZitadelClient{
			BaseURL:      cfg.IssuerURL,
			ServiceToken: cfg.ServiceToken,
			ProjectID:    cfg.ProjectID,
			HTTPClient: &http.Client{
				Timeout: timeout,
			},
		},
	}
}

// CreateUser creates a new user in Zitadel within the specified organization.
func (z *ZitadelClient) CreateUser(ctx context.Context, req CreateUserRequest) (*CreateUserResponse, error) {
	// Build the API request
	apiReq := models.ZitadelCreateUserRequestBody{
		UserName: req.Email,
		Password: req.Password,
	}
	apiReq.Profile.FirstName = req.FirstName
	apiReq.Profile.LastName = req.LastName
	apiReq.Email.Email = req.Email
	apiReq.Email.IsVerified = false

	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/management/v1/users/human/_import", z.ZitadelClient.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+z.ZitadelClient.ServiceToken)

	// Add org context if provided
	if req.OrgID != "" {
		httpReq.Header.Set("x-zitadel-orgid", req.OrgID)
	}

	resp, err := z.ZitadelClient.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		// Check for user already exists error
		if resp.StatusCode == http.StatusConflict {
			return nil, ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var apiResp models.ZitadelCreateUserResponseBody
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &CreateUserResponse{
		UserID: apiResp.UserID,
		Email:  req.Email,
	}, nil
}

// AssignUserRole assigns one or more roles to a user in the configured project.
func (z *ZitadelClient) AssignUserRole(ctx context.Context, userID string, roleKeys []string) error {
	apiReq := models.ZitadelRoleAssignment{
		RoleKeys: roleKeys,
	}

	body, err := json.Marshal(apiReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/management/v1/users/%s/grants", z.ZitadelClient.BaseURL, userID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+z.ZitadelClient.ServiceToken)

	resp, err := z.ZitadelClient.HTTPClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// AddUserToOrganization adds a user to an organization in Zitadel.
func (z *ZitadelClient) AddUserToOrganization(ctx context.Context, userID, orgID string) error {
	url := fmt.Sprintf("%s/management/v1/orgs/%s/members", z.ZitadelClient.BaseURL, orgID)

	reqBody := map[string]interface{}{
		"userId": userID,
		"roles":  []string{"ORG_OWNER"}, // Default role, can be customized
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+z.ZitadelClient.ServiceToken)

	resp, err := z.ZitadelClient.HTTPClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// DeactivateUser deactivates a user in Zitadel, preventing them from logging in.
func (z *ZitadelClient) DeactivateUser(ctx context.Context, userID string) error {
	url := fmt.Sprintf("%s/management/v1/users/%s/_deactivate", z.ZitadelClient.BaseURL, userID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+z.ZitadelClient.ServiceToken)

	resp, err := z.ZitadelClient.HTTPClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// ActivateUser activates a previously deactivated user in Zitadel.
func (z *ZitadelClient) ActivateUser(ctx context.Context, userID string) error {
	url := fmt.Sprintf("%s/management/v1/users/%s/_reactivate", z.ZitadelClient.BaseURL, userID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+z.ZitadelClient.ServiceToken)

	resp, err := z.ZitadelClient.HTTPClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// GetUser retrieves user details from Zitadel by user ID.
func (z *ZitadelClient) GetUser(ctx context.Context, userID string) (*UserResponse, error) {
	url := fmt.Sprintf("%s/management/v1/users/%s", z.ZitadelClient.BaseURL, userID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+z.ZitadelClient.ServiceToken)

	resp, err := z.ZitadelClient.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var apiResp models.ZitadelGetUserResponseBody
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &UserResponse{
		UserID:    apiResp.User.UserID,
		UserName:  apiResp.User.UserName,
		FirstName: apiResp.User.Human.Profile.FirstName,
		LastName:  apiResp.User.Human.Profile.LastName,
		Email:     apiResp.User.Human.Email.Email,
		State:     apiResp.User.State,
	}, nil
}
