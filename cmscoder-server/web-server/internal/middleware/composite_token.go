package middleware

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
)

// CompositeTokenFormat is the prefix for composite tokens.
const CompositeTokenFormat = "cmscoderv1_"

// ParseCompositeToken parses a composite token into its components.
// Format: cmscoderv1_<base64(modelApiKey:accessToken)>
// Returns (modelApiKey, accessToken, error).
func ParseCompositeToken(raw string) (modelApiKey, accessToken string, err error) {
	if !strings.HasPrefix(raw, CompositeTokenFormat) {
		return "", "", fmt.Errorf("invalid composite token format")
	}

	encoded := raw[len(CompositeTokenFormat):]
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", fmt.Errorf("invalid composite token encoding: %w", err)
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid composite token payload")
	}

	return parts[0], parts[1], nil
}

// RequireCompositeToken validates that the Authorization header contains a valid composite token.
// It extracts both the model API key and access token, storing them in context for downstream use.
// This middleware does NOT validate the tokens against user-service — it only parses the format.
// Downstream handlers should validate each token independently.
func RequireCompositeToken() ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		token := r.GetHeader("Authorization")
		if token == "" {
			writeModelAuthError(r, "authorization header is required")
			return
		}

		// Strip "Bearer " prefix if present.
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		modelApiKey, accessToken, err := ParseCompositeToken(token)
		if err != nil {
			writeModelAuthError(r, "invalid composite token")
			return
		}

		// Store parsed tokens in context for downstream validation.
		r.SetCtxVar("compositeModelApiKey", modelApiKey)
		r.SetCtxVar("compositeAccessToken", accessToken)

		r.Middleware.Next()
	}
}
