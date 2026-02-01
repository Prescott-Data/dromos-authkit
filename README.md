# Dromos AuthKit

[![License: Proprietary](https://img.shields.io/badge/License-Proprietary-red.svg)](LICENSE)

A lightweight, production-ready authentication and authorization middleware for [Gin](https://github.com/gin-gonic/gin)-based Dromos Suite APIs, using [Zitadel](https://zitadel.com/) OIDC tokens.


## Installation

```bash
go get github.com/Prescott-Data/dromos-authkit
```

## Quick Start

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/Prescott-Data/dromos-authkit"
)

func main() {
    r := gin.Default()

    // Configure authentication
    cfg := authkit.Config{
        IssuerURL: "https://your-zitadel-instance.com",
        Audience:  "your-project-id",
        SkipPaths: []string{"/health", "/metrics"},
    }

    // Apply authentication middleware
    r.Use(authkit.AuthN(cfg))

    // Protected route
    r.GET("/api/user", func(c *gin.Context) {
        userID := authkit.UserID(c)
        email := authkit.Email(c)

        c.JSON(200, gin.H{
            "user_id": userID,
            "email":   email,
        })
    })

    r.Run(":8080")
}
```

## Usage Examples

### Basic Authentication

The `AuthN` middleware validates JWT tokens and extracts claims:

```go
r.Use(authkit.AuthN(authkit.Config{
    IssuerURL: "https://zitadel.example.com",
    Audience:  "123456789@project_name",
}))
```

### Role-Based Authorization

Require specific roles for protected routes:

```go
// Require admin role
adminRoutes := r.Group("/admin")
adminRoutes.Use(authkit.RequireRole("admin"))
{
    adminRoutes.GET("/users", listUsers)
    adminRoutes.DELETE("/users/:id", deleteUser)
}

// Require any of multiple roles
editorRoutes := r.Group("/content")
editorRoutes.Use(authkit.RequireRole("editor", "admin"))
{
    editorRoutes.POST("/articles", createArticle)
}
```

### Accessing User Information

```go
r.GET("/profile", func(c *gin.Context) {
    // Get individual fields
    userID := authkit.UserID(c)
    email := authkit.Email(c)
    orgID := authkit.OrgID(c)

    // Get full claims
    claims := authkit.GetClaims(c)

    c.JSON(200, gin.H{
        "user_id": userID,
        "email":   email,
        "org_id":  orgID,
        "roles":   claims.Roles,
    })
})
```

### Custom Role Checks

```go
r.POST("/publish", func(c *gin.Context) {
    if !authkit.HasRole(c, "publisher") {
        c.JSON(403, gin.H{"error": "insufficient permissions"})
        return
    }

    // Process publication
    c.JSON(200, gin.H{"status": "published"})
})

// Check multiple roles
if authkit.HasAnyRole(c, "admin", "moderator") {
    // Allow access
}
```

### Multi-Tenant Applications

```go
r.GET("/api/data", func(c *gin.Context) {
    orgID := authkit.OrgID(c)

    // Filter data by organization
    data := fetchDataForOrg(orgID)

    c.JSON(200, data)
})
```

### Tenant-Scoped Middleware

```go
// Ensure user belongs to specific tenant
r.Use(authkit.RequireTenant("expected-org-id"))

// Or allow any tenant but store for later use
r.Use(authkit.AuthN(cfg))
r.GET("/data", func(c *gin.Context) {
    // OrgID is automatically extracted
    tenantID := authkit.OrgID(c)
    // ...
})
```

### Skip Authentication for Specific Routes

```go
cfg := authkit.Config{
    IssuerURL: "https://zitadel.example.com",
    Audience:  "project-id",
    SkipPaths: []string{
        "/health",
        "/metrics",
        "/api/v1/public/*",
    },
}

r.Use(authkit.AuthN(cfg))
```

### WebSocket Authentication

For WebSocket connections, tokens can be passed via query parameter:

```go
// Client: ws://localhost:8080/ws?token=eyJhbGc...

r.GET("/ws", func(c *gin.Context) {
    // Token automatically extracted from query param
    userID := authkit.UserID(c)

    // Upgrade to WebSocket
    upgrader.Upgrade(c.Writer, c.Request, nil)
})
```

## API Reference

### Configuration

```go
type Config struct {
    IssuerURL string   // Zitadel issuer URL
    Audience  string   // Expected audience (project ID)
    SkipPaths []string // Routes that bypass auth
}
```

### Middleware

- **`AuthN(cfg Config) gin.HandlerFunc`** - Authentication middleware
- **`RequireRole(roles ...string) gin.HandlerFunc`** - Authorization middleware
- **`RequireTenant(tenantID string) gin.HandlerFunc`** - Tenant validation middleware

### Claims Functions

- **`GetClaims(c *gin.Context) *Claims`** - Retrieve full claims object
- **`UserID(c *gin.Context) string`** - Get authenticated user ID
- **`Email(c *gin.Context) string`** - Get user email
- **`OrgID(c *gin.Context) string`** - Get organization ID
- **`HasRole(c *gin.Context, role string) bool`** - Check single role
- **`HasAnyRole(c *gin.Context, roles ...string) bool`** - Check multiple roles

### Claims Structure

```go
type Claims struct {
    Sub   string                 // User ID
    Email string                 // User email
    OrgID string                 // Organization ID
    Roles map[string]interface{} // Role assignments
}
```

## Configuration with Environment Variables

```go
import "os"

cfg := authkit.Config{
    IssuerURL: os.Getenv("ZITADEL_ISSUER_URL"),
    Audience:  os.Getenv("ZITADEL_AUDIENCE"),
    SkipPaths: []string{"/health"},
}
```

## Error Handling

The middleware returns standard HTTP error codes:

- **401 Unauthorized** - Missing or invalid token
- **403 Forbidden** - Valid token but insufficient permissions

```go
r.Use(func(c *gin.Context) {
    c.Next()

    if c.Writer.Status() == 401 {
        // Log authentication failure
    }
})
```

## Testing

```go
import (
    "testing"
    "net/http/httptest"
    "github.com/gin-gonic/gin"
)

func TestProtectedRoute(t *testing.T) {
    gin.SetMode(gin.TestMode)
    r := gin.New()

    // Mock or skip auth for tests
    cfg := authkit.Config{
        IssuerURL: "https://test.zitadel.com",
        Audience:  "test-project",
        SkipPaths: []string{"/test"},
    }

    r.Use(authkit.AuthN(cfg))
    r.GET("/test", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    w := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/test", nil)
    r.ServeHTTP(w, req)

    if w.Code != 200 {
        t.Errorf("expected 200, got %d", w.Code)
    }
}
```

## Performance Considerations

- **JWKS Caching**: Public keys are cached and refreshed automatically
- **Context Storage**: Claims are stored once per request in Gin context
- **Minimal Allocations**: Optimized for high-throughput scenarios

## Security Best Practices

1. **Always use HTTPS** in production
2. **Validate audience claims** to prevent token reuse across projects
3. **Keep skip paths minimal** - only exempt truly public endpoints
4. **Log authentication failures** for security monitoring
5. **Rotate Zitadel keys regularly**

## Zitadel Configuration

1. Create a new project in Zitadel
2. Note your Project ID (use as `Audience`)
3. Create roles within the project
4. Assign roles to users/organizations
5. Generate an API application for your service

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This software is proprietary and confidential. Copyright © 2026 Prescott Data. All rights reserved.

Unauthorized copying, distribution, modification, or use of this software is strictly prohibited. See the [LICENSE](LICENSE) file for complete terms.

## Support

- [Issue Tracker](https://github.com/Prescott-Data/dromos-authkit/issues)
- [Discussions](https://github.com/Prescott-Data/dromos-authkit/discussions)

## Acknowledgments

Built with:

- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [golang-jwt](https://github.com/golang-jwt/jwt)
- [Zitadel](https://zitadel.com/)
