package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// JWTClaims represents the standard JWT claims for Model Token.
type JWTClaims struct {
	Subject string `json:"sub"`   // User ID
	Session string `json:"sid"`   // Session ID
	Agent   string `json:"agent"` // Agent type (claude-code, opencode)
	Issued  int64  `json:"iat"`   // Issued at timestamp
	Expires int64  `json:"exp"`   // Expiration timestamp
}

// JWTHelper provides JWT signing and verification using HS256.
type JWTHelper struct {
	secret []byte
}

// NewJWTHelper creates a new JWT helper with the given secret.
func NewJWTHelper(secret string) *JWTHelper {
	return &JWTHelper{secret: []byte(secret)}
}

// GenerateToken creates a new JWT with the given claims and TTL.
func (j *JWTHelper) GenerateToken(userId, sessionId, agentType string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		Subject: userId,
		Session: sessionId,
		Agent:   agentType,
		Issued:  now.Unix(),
		Expires: now.Add(ttl).Unix(),
	}

	// Create header
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal header: %w", err)
	}

	// Create payload
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}

	// Base64 encode header and payload
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// Create signature
	signingInput := headerB64 + "." + payloadB64
	signature := j.sign(signingInput)
	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)

	// Combine all parts
	token := signingInput + "." + signatureB64
	return token, nil
}

// VerifyToken validates a JWT and returns the claims if valid.
func (j *JWTHelper) VerifyToken(token string) (*JWTClaims, error) {
	// Split token into parts
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	headerB64, payloadB64, signatureB64 := parts[0], parts[1], parts[2]

	// Verify signature
	signingInput := headerB64 + "." + payloadB64
	expectedSignature := j.sign(signingInput)
	providedSignature, err := base64.RawURLEncoding.DecodeString(signatureB64)
	if err != nil {
		return nil, fmt.Errorf("invalid signature encoding: %w", err)
	}

	if !hmac.Equal(expectedSignature, providedSignature) {
		return nil, fmt.Errorf("invalid signature")
	}

	// Decode payload
	payloadJSON, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return nil, fmt.Errorf("invalid payload encoding: %w", err)
	}

	var claims JWTClaims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	// Check expiration
	if time.Now().Unix() > claims.Expires {
		return nil, fmt.Errorf("token has expired")
	}

	return &claims, nil
}

// sign creates an HMAC-SHA256 signature of the input.
func (j *JWTHelper) sign(input string) []byte {
	h := hmac.New(sha256.New, j.secret)
	h.Write([]byte(input))
	return h.Sum(nil)
}

// VerifyHMAC validates an HMAC-SHA256 signature.
// Returns true if the provided signature matches the expected signature.
func VerifyHMAC(message, secret, providedSignature string) bool {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	expectedSignature := hexEncode(h.Sum(nil))
	return hmac.Equal([]byte(expectedSignature), []byte(providedSignature))
}

// hexEncode converts bytes to lowercase hex string.
func hexEncode(data []byte) string {
	const hexTable = "0123456789abcdef"
	result := make([]byte, len(data)*2)
	for i, b := range data {
		result[i*2] = hexTable[b>>4]
		result[i*2+1] = hexTable[b&0x0f]
	}
	return string(result)
}
