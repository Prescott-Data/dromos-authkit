# Basic Example

This example demonstrates the basic usage of dromos-authkit with a Gin application.

## Features Demonstrated

- Global authentication middleware
- Public routes (health checks)
- Protected routes requiring authentication
- Role-based authorization (admin, editor)
- Custom role checks
- Accessing user information

## Setup

1. **Set environment variables**:

```bash
export ZITADEL_ISSUER_URL="https://your-zitadel-instance.com"
export ZITADEL_AUDIENCE="your-project-id"
export PORT="3000"
```

2. **Run the example**:

```bash
go run main.go
```

## Testing

### Public Endpoints

```bash
# Health check (no auth required)
curl http://localhost:3000/health

# Public endpoint (no auth required)
curl http://localhost:3000/public
```

### Protected Endpoints

```bash
# Get current user info (requires valid JWT)
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:3000/api/me

# Get full claims
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:3000/api/claims
```

### Role-Based Endpoints

```bash
# Admin endpoint (requires 'admin' role)
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:3000/api/admin/users

# Editor endpoint (requires 'editor' or 'admin' role)
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:3000/api/content/articles

# Publish endpoint (requires 'publisher' or 'admin' role)
curl -X POST \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:3000/api/publish
```

## Expected Responses

### Authenticated User Info

```json
{
  "user_id": "123456789",
  "email": "user@example.com",
  "org_id": "987654321"
}
```

### Unauthorized (401)

```json
{
  "error": "missing or invalid Authorization header"
}
```

### Forbidden (403)

```json
{
  "error": "insufficient permissions — requires one of: admin"
}
```

## Getting a Test Token

To get a valid JWT token from Zitadel:

1. Create a user in your Zitadel instance
2. Assign appropriate roles to the user
3. Use Zitadel's OAuth2/OIDC flow to obtain an access token
4. Use the token in the Authorization header

Example using curl to get a token (machine-to-machine):

```bash
curl -X POST https://your-zitadel-instance.com/oauth/v2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET" \
  -d "scope=openid profile email"
```
