package authkit

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"
)

// jwksKey represents a single key from the JWKS endpoint.
type jwksKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// jwksResponse represents the JWKS endpoint response.
type jwksResponse struct {
	Keys []jwksKey `json:"keys"`
}

// JWKSCache fetches and caches JWKS keys from the identity provider.
type JWKSCache struct {
	jwksURL    string
	keys       map[string]*rsa.PublicKey
	mu         sync.RWMutex
	lastFetch  time.Time
	cacheTTL   time.Duration
	httpClient *http.Client
}

// NewJWKSCache creates a new JWKS cache for the given URL.
func NewJWKSCache(jwksURL string) *JWKSCache {
	return &JWKSCache{
		jwksURL:  jwksURL,
		keys:     make(map[string]*rsa.PublicKey),
		cacheTTL: 1 * time.Hour,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetKey returns the RSA public key for the given key ID.
// It fetches fresh keys if the cache is stale or the key ID is unknown.
func (j *JWKSCache) GetKey(kid string) (*rsa.PublicKey, error) {
	// Try cached key first
	j.mu.RLock()
	if key, ok := j.keys[kid]; ok && time.Since(j.lastFetch) < j.cacheTTL {
		j.mu.RUnlock()
		return key, nil
	}
	j.mu.RUnlock()

	// Fetch fresh keys
	if err := j.refresh(); err != nil {
		return nil, fmt.Errorf("failed to refresh JWKS: %w", err)
	}

	j.mu.RLock()
	defer j.mu.RUnlock()
	key, ok := j.keys[kid]
	if !ok {
		return nil, fmt.Errorf("key %q not found in JWKS", kid)
	}
	return key, nil
}

func (j *JWKSCache) refresh() error {
	j.mu.Lock()
	defer j.mu.Unlock()

	// Double-check after acquiring write lock
	if time.Since(j.lastFetch) < 30*time.Second {
		return nil
	}

	resp, err := j.httpClient.Get(j.jwksURL)
	if err != nil {
		return fmt.Errorf("JWKS fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS: %w", err)
	}

	newKeys := make(map[string]*rsa.PublicKey, len(jwks.Keys))
	for _, k := range jwks.Keys {
		if k.Kty != "RSA" || k.Use != "sig" {
			continue
		}
		pubKey, err := parseRSAPublicKey(k.N, k.E)
		if err != nil {
			continue
		}
		newKeys[k.Kid] = pubKey
	}

	j.keys = newKeys
	j.lastFetch = time.Now()
	return nil
}

func parseRSAPublicKey(nStr, eStr string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, fmt.Errorf("invalid modulus: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, fmt.Errorf("invalid exponent: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}

	return &rsa.PublicKey{N: n, E: e}, nil
}
