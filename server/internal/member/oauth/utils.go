package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

func absoluteURL(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && parsed.Host != "" && (parsed.Scheme == "http" || parsed.Scheme == "https")
}

func token(bytes int) (string, error) {
	buffer := make([]byte, bytes)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func challenge(verifier string) string {
	digest := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(digest[:])
}

func idTokenClaims(value string) map[string]any {
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return nil
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil
	}

	claims := map[string]any{}
	if json.Unmarshal(payload, &claims) != nil {
		return nil
	}

	return claims
}

func claim(claims map[string]any, field string) string {
	if value, ok := claims[field]; ok {
		return fmt.Sprint(value)
	}
	return ""
}

func boolClaim(claims map[string]any, field string) bool {
	value, ok := claims[field]
	if !ok {
		return false
	}
	switch value := value.(type) {
	case bool:
		return value
	case string:
		return strings.EqualFold(value, "true")
	case float64:
		return value != 0
	}
	return false
}

func administrator(email string, isAdmin bool, emails []string) bool {
	if isAdmin {
		return true
	}
	for _, candidate := range emails {
		if email != "" && strings.EqualFold(strings.TrimSpace(email), strings.TrimSpace(candidate)) {
			return true
		}
	}
	return false
}
