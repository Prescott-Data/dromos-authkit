package authkit

import (
	"bytes"
	"context"
	"encoding/base64"
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

// OrganizationResponse is an alias to models.OrganizationResponse for backward compatibility.
type OrganizationResponse = models.OrganizationResponse

// IDPLink is an alias to models.IDPLink for backward compatibility.
type IDPLink = models.IDPLink

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
	// Use provided username, fallback to email if username is empty
	username := req.UserName
	if username == "" {
		username = req.Email
	}

	apiReq := models.ZitadelCreateUserRequestBody{
		UserName: username,
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
		ProjectID: z.ZitadelClient.ProjectID,
		RoleKeys:  roleKeys,
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
		AvatarURL: apiResp.User.Human.Profile.AvatarURL,
	}, nil
}

// GetOrganization retrieves organization details from Zitadel by organization ID.
func (z *ZitadelClient) GetOrganization(ctx context.Context, orgID string) (*OrganizationResponse, error) {
	url := fmt.Sprintf("%s/management/v1/orgs/me", z.ZitadelClient.BaseURL)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+z.ZitadelClient.ServiceToken)
	httpReq.Header.Set("x-zitadel-orgid", orgID)

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

	var apiResp models.ZitadelGetOrgResponseBody
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	org := &OrganizationResponse{
		ID:     apiResp.Org.ID,
		Name:   apiResp.Org.Name,
		State:  apiResp.Org.State,
		Domain: apiResp.Org.PrimaryDomain,
	}

	// Fetch label policy for org branding (logo, icon)
	labelPolicy, err := z.GetOrgLabelPolicy(ctx, orgID)
	if err == nil && labelPolicy != nil {
		org.LogoURL = labelPolicy.LogoURL
		org.LogoDarkURL = labelPolicy.LogoDarkURL
		org.IconURL = labelPolicy.IconURL
		org.IconDarkURL = labelPolicy.IconDarkURL
	}

	return org, nil
}

// LabelPolicy contains the branding/label policy for an organization
type LabelPolicy struct {
	LogoURL         string
	LogoDarkURL     string
	IconURL         string
	IconDarkURL     string
	PrimaryColor    string
	BackgroundColor string
}

// GetOrgLabelPolicy retrieves the label/branding policy for an organization.
func (z *ZitadelClient) GetOrgLabelPolicy(ctx context.Context, orgID string) (*LabelPolicy, error) {
	url := fmt.Sprintf("%s/management/v1/policies/label", z.ZitadelClient.BaseURL)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+z.ZitadelClient.ServiceToken)
	httpReq.Header.Set("x-zitadel-orgid", orgID)

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

	var apiResp models.ZitadelLabelPolicyResponseBody
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &LabelPolicy{
		LogoURL:         apiResp.Policy.LogoURL,
		LogoDarkURL:     apiResp.Policy.LogoDarkURL,
		IconURL:         apiResp.Policy.IconURL,
		IconDarkURL:     apiResp.Policy.IconDarkURL,
		PrimaryColor:    apiResp.Policy.PrimaryColor,
		BackgroundColor: apiResp.Policy.BackgroundColor,
	}, nil
}

// ListProjectUserGrants lists all user grants for the configured project in an organization.
func (z *ZitadelClient) ListProjectUserGrants(ctx context.Context, orgID string) ([]models.UserGrant, error) {
	url := fmt.Sprintf("%s/management/v1/projects/%s/grants/_search", z.ZitadelClient.BaseURL, z.ZitadelClient.ProjectID)

	// Empty search body to get all grants
	reqBody := map[string]any{}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+z.ZitadelClient.ServiceToken)
	httpReq.Header.Set("x-zitadel-orgid", orgID)

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

	var apiResp models.ZitadelListUserGrantsResponseBody
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	grants := make([]models.UserGrant, 0, len(apiResp.Result))
	for _, g := range apiResp.Result {
		grants = append(grants, models.UserGrant{
			ID:        g.ID,
			UserID:    g.UserID,
			ProjectID: g.ProjectID,
			RoleKeys:  g.RoleKeys,
			State:     g.State,
			UserName:  g.UserName,
			FirstName: g.FirstName,
			LastName:  g.LastName,
			Email:     g.Email,
			AvatarURL: g.AvatarURL,
		})
	}

	return grants, nil
}

// ListUserGrantsInOrg lists all user grants in an organization for the configured project.
func (z *ZitadelClient) ListUserGrantsInOrg(ctx context.Context, orgID string) ([]models.UserGrant, error) {
	url := fmt.Sprintf("%s/management/v1/users/grants/_search", z.ZitadelClient.BaseURL)

	reqBody := map[string]any{
		"queries": []map[string]any{
			{
				"projectIdQuery": map[string]string{
					"projectId": z.ZitadelClient.ProjectID,
				},
			},
		},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+z.ZitadelClient.ServiceToken)
	httpReq.Header.Set("x-zitadel-orgid", orgID)

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

	var apiResp models.ZitadelListUserGrantsResponseBody
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	grants := make([]models.UserGrant, 0, len(apiResp.Result))
	for _, g := range apiResp.Result {
		grants = append(grants, models.UserGrant{
			ID:        g.ID,
			UserID:    g.UserID,
			ProjectID: g.ProjectID,
			RoleKeys:  g.RoleKeys,
			State:     g.State,
			UserName:  g.UserName,
			FirstName: g.FirstName,
			LastName:  g.LastName,
			Email:     g.Email,
			AvatarURL: g.AvatarURL,
		})
	}

	return grants, nil
}

// GetOrgMetadata retrieves all metadata for an organization.
func (z *ZitadelClient) GetOrgMetadata(ctx context.Context, orgID string) (map[string]string, error) {
	url := fmt.Sprintf("%s/management/v1/metadata/_search", z.ZitadelClient.BaseURL)

	reqBody := map[string]any{}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+z.ZitadelClient.ServiceToken)
	httpReq.Header.Set("x-zitadel-orgid", orgID)

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

	var apiResp models.ZitadelListOrgMetadataResponseBody
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	metadata := make(map[string]string)
	for _, m := range apiResp.Result {
		// Zitadel returns base64-encoded values
		decoded, err := base64.StdEncoding.DecodeString(m.Value)
		if err != nil {
			metadata[m.Key] = m.Value // Use raw value if decode fails
		} else {
			metadata[m.Key] = string(decoded)
		}
	}

	return metadata, nil
}

// SetOrgMetadata sets a metadata key-value pair for an organization.
func (z *ZitadelClient) SetOrgMetadata(ctx context.Context, orgID, key, value string) error {
	url := fmt.Sprintf("%s/management/v1/metadata/%s", z.ZitadelClient.BaseURL, key)

	// Zitadel expects base64-encoded value
	encodedValue := base64.StdEncoding.EncodeToString([]byte(value))
	reqBody := map[string]string{
		"value": encodedValue,
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
	httpReq.Header.Set("x-zitadel-orgid", orgID)

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

// SearchUserByEmail searches for a user by email address.
func (z *ZitadelClient) SearchUserByEmail(ctx context.Context, email string) (*UserResponse, error) {
	url := fmt.Sprintf("%s/v2/users", z.ZitadelClient.BaseURL)

	reqBody := map[string]any{
		"queries": []map[string]any{
			{
				"emailQuery": map[string]any{
					"emailAddress": email,
					"method":       "TEXT_QUERY_METHOD_EQUALS",
				},
			},
		},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
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

	var apiResp models.ZitadelSearchUsersResponseBody
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(apiResp.Result) == 0 {
		return nil, nil // User not found
	}

	user := apiResp.Result[0]
	userResp := &UserResponse{
		UserID:   user.UserID,
		UserName: user.UserName,
		State:    user.State,
	}

	if user.Human != nil {
		userResp.FirstName = user.Human.Profile.FirstName
		userResp.LastName = user.Human.Profile.LastName
		userResp.Email = user.Human.Email.Email
	}

	return userResp, nil
}

// GetUserGrantForProject checks if a user has a grant for the configured project.
func (z *ZitadelClient) GetUserGrantForProject(ctx context.Context, userID string) (*models.UserGrant, error) {
	// Use the user grants search endpoint with project filter
	url := fmt.Sprintf("%s/management/v1/users/grants/_search", z.ZitadelClient.BaseURL)

	reqBody := map[string]any{
		"queries": []map[string]any{
			{
				"userIdQuery": map[string]string{
					"userId": userID,
				},
			},
			{
				"projectIdQuery": map[string]string{
					"projectId": z.ZitadelClient.ProjectID,
				},
			},
		},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
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

	var apiResp models.ZitadelListUserGrantsResponseBody
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(apiResp.Result) == 0 {
		return nil, nil // No grant found for this project
	}

	g := apiResp.Result[0]
	return &models.UserGrant{
		ID:        g.ID,
		UserID:    g.UserID,
		ProjectID: g.ProjectID,
		RoleKeys:  g.RoleKeys,
		State:     g.State,
		UserName:  g.UserName,
		FirstName: g.FirstName,
		LastName:  g.LastName,
		Email:     g.Email,
		AvatarURL: g.AvatarURL,
	}, nil
}

// GetUserIDPLinks retrieves all external identity provider links for a user.
func (z *ZitadelClient) GetUserIDPLinks(ctx context.Context, userID string) ([]IDPLink, error) {
	url := fmt.Sprintf("%s/management/v1/users/%s/idps/_search", z.ZitadelClient.BaseURL, userID)

	// Empty search body to get all links
	reqBody := map[string]any{}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
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

	var apiResp models.ZitadelListIDPLinksResponseBody
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	links := make([]IDPLink, 0, len(apiResp.Result))
	for _, link := range apiResp.Result {
		links = append(links, IDPLink{
			IDPID:          link.IDPID,
			IDPName:        link.IDPName,
			UserID:         link.UserID,
			ExternalUserID: link.ProvidedUserID,
			ProvidedUserID: link.ProvidedUserID,
			ProvidedEmail:  link.ProvidedEmail,
		})
	}

	return links, nil
}

// RemoveUserIDPLink removes an external identity provider link from a user.
func (z *ZitadelClient) RemoveUserIDPLink(ctx context.Context, userID, idpID, externalUserID string) error {
	url := fmt.Sprintf("%s/management/v1/users/%s/idps/%s/%s", z.ZitadelClient.BaseURL, userID, idpID, externalUserID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
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
