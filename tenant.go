package authkit

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireTenant returns a Gin middleware that ensures the authenticated user
// has an organization context (org_id). This must be applied AFTER AuthN.
// Requests without an org_id are rejected with 403 Forbidden.
func RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID := OrgID(c)
		if orgID == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "no organization context — user must belong to an organization",
			})
			return
		}
		c.Next()
	}
}
