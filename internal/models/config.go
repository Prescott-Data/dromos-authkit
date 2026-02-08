package models

// Config holds the configuration for the auth middleware.
type Config struct {
	IssuerURL string
	Audience string
	SkipPaths []string
}
