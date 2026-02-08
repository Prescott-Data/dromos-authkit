package authkit

import (
	"fmt"
	"net/http"

	"github.com/Prescott-Data/dromos-authkit/internal/models"
	"github.com/gin-gonic/gin"
)

// OrgRole is an alias to models.OrgRole for backward compatibility.
type OrgRole = models.OrgRole

// Organization role constants matching Zitadel role keys.
const (
	// OrgRoleOwner has full control over the organization.
	OrgRoleOwner = models.OrgRoleOwner

	// OrgRoleAdmin can manage members and most organization settings.
	OrgRoleAdmin = models.OrgRoleAdmin

	// OrgRoleMember has standard access to organization resources.
	OrgRoleMember = models.OrgRoleMember

	// OrgRoleViewer has read-only access to organization resources.
	OrgRoleViewer = models.OrgRoleViewer
)

// RequireOrgRole returns a Gin middleware that checks if the authenticated user
// has at least one of the specified organization roles. Must be applied AFTER AuthN.
func RequireOrgRole(roles ...OrgRole) gin.HandlerFunc {
	roleStrs := make([]string, len(roles))
	for i, r := range roles {
		roleStrs[i] = string(r)
	}

	return func(c *gin.Context) {
		if HasAnyRole(c, roleStrs...) {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": fmt.Sprintf("insufficient permissions — requires one of: %s", formatRoles(roles)),
		})
	}
}

// IsOrgAdmin checks if the authenticated user has admin or owner privileges.
func IsOrgAdmin(c *gin.Context) bool {
	return HasAnyRole(c, string(OrgRoleOwner), string(OrgRoleAdmin))
}

// IsOrgOwner checks if the authenticated user is an organization owner.
func IsOrgOwner(c *gin.Context) bool {
	return HasRole(c, string(OrgRoleOwner))
}

// CanManageMembers checks if the authenticated user can invite and manage members.
func CanManageMembers(c *gin.Context) bool {
	return IsOrgAdmin(c)
}

// HasAnyOrgRole checks if the authenticated user has any of the specified organization roles.
func HasAnyOrgRole(c *gin.Context, roles ...OrgRole) bool {
	roleStrs := make([]string, len(roles))
	for i, r := range roles {
		roleStrs[i] = string(r)
	}
	return HasAnyRole(c, roleStrs...)
}

// formatRoles formats a slice of OrgRoles into a comma-separated string.
func formatRoles(roles []OrgRole) string {
	if len(roles) == 0 {
		return ""
	}
	if len(roles) == 1 {
		return string(roles[0])
	}

	result := string(roles[0])
	for i := 1; i < len(roles); i++ {
		result += ", " + string(roles[i])
	}
	return result
}
