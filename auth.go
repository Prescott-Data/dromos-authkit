package authkit

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthN returns a Gin middleware that validates Zitadel JWT access tokens.
// It extracts the Bearer token from the Authorization header (or "token" query
// parameter for WebSocket upgrades), validates it against the JWKS endpoint,
// and stores the parsed claims in the Gin context.
func AuthN(cfg Config) gin.HandlerFunc {
	jwks := NewJWKSCache(cfg.IssuerURL + "/oauth/v2/keys")

	skipSet := make(map[string]bool, len(cfg.SkipPaths))
	for _, p := range cfg.SkipPaths {
		skipSet[p] = true
	}

	log.Printf("[authkit] Initialized AuthN middleware (issuer=%s, audience=%s, skip=%d paths)",
		cfg.IssuerURL, cfg.Audience, len(cfg.SkipPaths))

	return func(c *gin.Context) {
		// Skip configured paths
		if skipSet[c.FullPath()] {
			c.Next()
			return
		}

		// Extract token from Authorization header or query param (WebSocket fallback)
		tokenStr := extractToken(c)
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing or invalid Authorization header",
			})
			return
		}

		// Parse and validate the JWT
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// Get the key ID from the token header
			kid, ok := token.Header["kid"].(string)
			if !ok {
				return nil, fmt.Errorf("missing kid in token header")
			}

			// Fetch the public key from JWKS cache
			key, err := jwks.GetKey(kid)
			if err != nil {
				return nil, err
			}
			return key, nil
		},
			jwt.WithIssuer(cfg.IssuerURL),
			jwt.WithValidMethods([]string{"RS256"}),
		)

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		// Also validate audience if configured
		if cfg.Audience != "" {
			if err := validateAudience(token, cfg.Audience); err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "token audience mismatch",
				})
				return
			}
		}

		// Extract claims into our struct
		mapClaims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token claims",
			})
			return
		}

		claims := &Claims{
			Sub:   getStringClaim(mapClaims, "sub"),
			Email: getStringClaim(mapClaims, "email"),
			OrgID: getStringClaim(mapClaims, "urn:zitadel:iam:org:id"),
		}

		// Extract project roles
		if roles, ok := mapClaims["urn:zitadel:iam:org:project:roles"].(map[string]interface{}); ok {
			claims.Roles = roles
		}

		SetClaims(c, claims)
		c.Next()
	}
}

// extractToken gets the JWT from the Authorization header or "token" query param.
func extractToken(c *gin.Context) string {
	// Try Authorization header first
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Fallback: query parameter (for WebSocket connections where browsers
	// cannot set custom headers on the upgrade request)
	if token := c.Query("token"); token != "" {
		return token
	}

	return ""
}

func validateAudience(token *jwt.Token, expectedAudience string) error {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("invalid claims type")
	}

	// Zitadel may include audience as a string or array
	switch aud := claims["aud"].(type) {
	case string:
		if aud == expectedAudience {
			return nil
		}
	case []interface{}:
		for _, a := range aud {
			if s, ok := a.(string); ok && s == expectedAudience {
				return nil
			}
		}
	}

	return fmt.Errorf("audience %q not found in token", expectedAudience)
}

// getStringClaim safely extracts a string claim from JWT MapClaims.
func getStringClaim(m jwt.MapClaims, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// KeyFunc returns a jwt.Keyfunc backed by the JWKS cache.
// This is useful for external code that needs to validate tokens directly.
func KeyFunc(jwks *JWKSCache) jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing kid in token header")
		}
		return jwks.GetKey(kid)
	}
}

// ValidateToken validates a raw JWT string and returns the claims.
// Useful for validating tokens outside of HTTP middleware (e.g. WebSocket re-auth).
func ValidateToken(tokenStr string, cfg Config) (*Claims, error) {
	jwks := NewJWKSCache(cfg.IssuerURL + "/oauth/v2/keys")

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing kid in token header")
		}
		key, err := jwks.GetKey(kid)
		if err != nil {
			return nil, err
		}
		return key, nil
	},
		jwt.WithIssuer(cfg.IssuerURL),
		jwt.WithValidMethods([]string{"RS256"}),
	)

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type")
	}

	claims := &Claims{
		Sub:   getStringClaim(mapClaims, "sub"),
		Email: getStringClaim(mapClaims, "email"),
		OrgID: getStringClaim(mapClaims, "urn:zitadel:iam:org:id"),
	}
	if roles, ok := mapClaims["urn:zitadel:iam:org:project:roles"].(map[string]interface{}); ok {
		claims.Roles = roles
	}

	return claims, nil
}
