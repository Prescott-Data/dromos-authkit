package authkit

import (
	"github.com/gin-gonic/gin"
)

const claimsKey = "dromos_auth_claims"

// Claims represents the validated JWT claims from Zitadel.
type Claims struct {
	// Sub is the Zitadel user ID.
	Sub string `json:"sub"`

	// Email is the user's email address.
	Email string `json:"email"`

	// OrgID is the Zitadel organization ID the user belongs to.
	OrgID string `json:"urn:zitadel:iam:org:id"`

	// OrgDomain is the primary domain of the user's resource owner organization.
	OrgDomain string `json:"urn:zitadel:iam:user:resourceowner:primary_domain"`

	// Roles maps role names to their grant details.
	// The keys are role names (e.g. "admin", "editor").
	Roles map[string]interface{} `json:"urn:zitadel:iam:org:project:roles"`
}

// SetClaims stores validated claims in the Gin context.
func SetClaims(c *gin.Context, claims *Claims) {
	c.Set(claimsKey, claims)
}

// GetClaims retrieves validated claims from the Gin context.
// Returns nil if no claims are set (unauthenticated request).
func GetClaims(c *gin.Context) *Claims {
	val, exists := c.Get(claimsKey)
	if !exists {
		return nil
	}
	claims, ok := val.(*Claims)
	if !ok {
		return nil
	}
	return claims
}

// UserID returns the authenticated user's Zitadel subject ID.
// Returns empty string if the request is not authenticated.
func UserID(c *gin.Context) string {
	if cl := GetClaims(c); cl != nil {
		return cl.Sub
	}
	return ""
}

// OrgID returns the authenticated user's organization ID.
// Returns empty string if no org context is available.
func OrgID(c *gin.Context) string {
	if cl := GetClaims(c); cl != nil {
		return cl.OrgID
	}
	return ""
}

// OrgDomain returns the authenticated user's organization primary domain.
// Returns empty string if no org domain is available.
func OrgDomain(c *gin.Context) string {
	if cl := GetClaims(c); cl != nil {
		return cl.OrgDomain
	}
	return ""
}

// Email returns the authenticated user's email.
func Email(c *gin.Context) string {
	if cl := GetClaims(c); cl != nil {
		return cl.Email
	}
	return ""
}

// HasRole checks if the authenticated user has the specified role.
func HasRole(c *gin.Context, role string) bool {
	cl := GetClaims(c)
	if cl == nil || cl.Roles == nil {
		return false
	}
	_, ok := cl.Roles[role]
	return ok
}

// HasAnyRole checks if the authenticated user has at least one of the specified roles.
func HasAnyRole(c *gin.Context, roles ...string) bool {
	for _, role := range roles {
		if HasRole(c, role) {
			return true
		}
	}
	return false
}
