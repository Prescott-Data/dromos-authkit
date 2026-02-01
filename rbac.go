package authkit

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// RequireRole returns a Gin middleware that checks if the authenticated user
// has at least one of the specified roles. Must be applied AFTER AuthN.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if HasAnyRole(c, roles...) {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": fmt.Sprintf("insufficient permissions — requires one of: %s", strings.Join(roles, ", ")),
		})
	}
}
