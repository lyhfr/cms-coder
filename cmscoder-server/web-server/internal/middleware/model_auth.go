package middleware

import (
	"cmscoder-web-server/internal/clients/userclient"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// ModelAuth validates the composite token from the Authorization header.
// It extracts both the model API key and access token from the composite token,
// validates the model key via user-service, and verifies that both tokens
// belong to the same active session.
func ModelAuth(userClient *userclient.Client) ghttp.HandlerFunc {
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

		// Parse composite token.
		modelApiKey, accessToken, err := ParseCompositeToken(token)
		if err != nil {
			writeModelAuthError(r, "invalid composite token")
			return
		}

		// Validate model API key via user-service.
		result, err := userClient.ValidateModelKey(r.GetCtx(), modelApiKey)
		if err != nil {
			writeModelAuthError(r, "invalid model api key")
			return
		}

		// CRITICAL: Verify the model API key's sessionId matches the access token.
		// This ensures both tokens belong to the same session, preventing abuse
		// where a leaked model key is paired with an unrelated access token.
		if result.SessionId != accessToken {
			writeModelAuthError(r, "token mismatch: model key and access token do not belong to the same session")
			return
		}

		// Verify the session is still active.
		_, err = userClient.IntrospectSession(r.GetCtx(), accessToken)
		if err != nil {
			writeModelAuthError(r, "associated session has expired or been revoked")
			return
		}

		// Store validated info in context for downstream use.
		r.SetCtxVar("modelApiKey", modelApiKey)
		r.SetCtxVar("modelKeyUserId", result.UserId)
		r.SetCtxVar("modelKeySessionId", result.SessionId)
		r.SetCtxVar("modelKeyAgentType", result.AgentType)
		r.SetCtxVar("modelKeyPluginInstance", result.PluginInstance)

		r.Middleware.Next()
	}
}

// writeModelAuthError writes a 401 unauthorized response for model endpoints.
func writeModelAuthError(r *ghttp.Request, message string) {
	r.Response.WriteStatus(401)
	r.Response.WriteJson(ghttp.DefaultHandlerResponse{
		Code:    gcode.CodeNotAuthorized.Code(),
		Message: message,
	})
	r.Exit()
}

// NewModelAuthError creates a model auth error using gerror.
func NewModelAuthError(message string) error {
	return gerror.NewCode(gcode.New(401, message, nil), message)
}
