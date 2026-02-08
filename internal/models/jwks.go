package models

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

// JWKSKey represents a single key from the JWKS endpoint.
type JWKSKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKSResponse represents the JWKS endpoint response.
type JWKSResponse struct {
	Keys []JWKSKey `json:"keys"`
}

// JWKSCache fetches and caches JWKS keys from the identity provider.
type JWKSCache struct {
	JWKSURL    string
	Keys       map[string]*rsa.PublicKey
	Mu         sync.RWMutex
	LastFetch  time.Time
	CacheTTL   time.Duration
	HTTPClient *http.Client
}

// GetKey returns the RSA public key for the given key ID.
func (j *JWKSCache) GetKey(kid string) (*rsa.PublicKey, error) {
	// Try cached key first
	j.Mu.RLock()
	if key, ok := j.Keys[kid]; ok && time.Since(j.LastFetch) < j.CacheTTL {
		j.Mu.RUnlock()
		return key, nil
	}
	j.Mu.RUnlock()

	// Fetch fresh keys
	if err := j.refresh(); err != nil {
		return nil, fmt.Errorf("failed to refresh JWKS: %w", err)
	}

	j.Mu.RLock()
	defer j.Mu.RUnlock()
	key, ok := j.Keys[kid]
	if !ok {
		return nil, fmt.Errorf("key %q not found in JWKS", kid)
	}
	return key, nil
}

func (j *JWKSCache) refresh() error {
	j.Mu.Lock()
	defer j.Mu.Unlock()

	// Double-check after acquiring write lock
	if time.Since(j.LastFetch) < 30*time.Second {
		return nil
	}

	resp, err := j.HTTPClient.Get(j.JWKSURL)
	if err != nil {
		return fmt.Errorf("JWKS fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	var jwks JWKSResponse
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

	j.Keys = newKeys
	j.LastFetch = time.Now()
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
