package authkit

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/Prescott-Data/dromos-authkit/internal/models"
)

// InvitationClaims is an alias to models.InvitationClaims for backward compatibility.
type InvitationClaims = models.InvitationClaims

// AccessCode is an alias to models.AccessCode for backward compatibility.
type AccessCode = models.AccessCode

// GenerateAccessCode creates a cryptographically secure access code
func GenerateAccessCode() (string, error) {
	// Character set excluding ambiguous characters: 0, O, 1, I, L
	const charset = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"
	const codeLength = 12

	// Generate random bytes
	randomBytes := make([]byte, codeLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Map random bytes to charset
	code := make([]byte, codeLength)
	for i, b := range randomBytes {
		code[i] = charset[int(b)%len(charset)]
	}

	// Format as XXXX-XXXX-XXXX
	formatted := fmt.Sprintf("%s-%s-%s",
		string(code[0:4]),
		string(code[4:8]),
		string(code[8:12]),
	)

	return formatted, nil
}

// ValidateAccessCodeFormat checks if the access code matches the expected format
func ValidateAccessCodeFormat(code string) bool {
	// Check length (14 chars: 12 alphanumeric + 2 hyphens)
	if len(code) != 14 {
		return false
	}

	// Check format: XXXX-XXXX-XXXX
	parts := strings.Split(code, "-")
	if len(parts) != 3 {
		return false
	}

	// Each part must be exactly 4 characters
	for _, part := range parts {
		if len(part) != 4 {
			return false
		}
		// Validate characters (uppercase letters and digits, excluding ambiguous ones)
		for _, ch := range part {
			if !isValidAccessCodeChar(ch) {
				return false
			}
		}
	}

	return true
}

// isValidAccessCodeChar checks if a character is valid for access codes.
func isValidAccessCodeChar(ch rune) bool {
	const validChars = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"
	return strings.ContainsRune(validChars, ch)
}

// HashAccessCode creates a SHA256 hash of the access code for secure storage.
func HashAccessCode(code string) string {
	hash := sha256.Sum256([]byte(code))
	return hex.EncodeToString(hash[:])
}
