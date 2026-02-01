package authkit

// Config holds the configuration for the auth middleware.
type Config struct {
	// IssuerURL is the Zitadel issuer URL (e.g. "http://172.191.51.250:8080").
	IssuerURL string

	// Audience is the expected audience claim (Zitadel project ID).
	Audience string

	// SkipPaths lists route paths that bypass authentication (e.g. health checks).
	// These should match Gin's FullPath() patterns (e.g. "/api/v1/health").
	SkipPaths []string
}
