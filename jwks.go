package authkit

import (
	"crypto/rsa"
	"net/http"
	"time"

	"github.com/Prescott-Data/dromos-authkit/internal/models"
)

// JWKSCache is an alias to models.JWKSCache for backward compatibility.
type JWKSCache = models.JWKSCache

// NewJWKSCache creates a new JWKS cache for the given URL.
func NewJWKSCache(jwksURL string) *JWKSCache {
	return &JWKSCache{
		JWKSURL:  jwksURL,
		Keys:     make(map[string]*rsa.PublicKey),
		CacheTTL: 1 * time.Hour,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}
