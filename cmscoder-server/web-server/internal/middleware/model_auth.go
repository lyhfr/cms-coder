package middleware

import (
	"cmscoder-web-server/internal/clients/userclient"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// ModelAuth validates the model API key from the Authorization header.
// It extracts the Bearer token, validates it as a model API key via user-service,
// and checks that the associated session is still active.
func ModelAuth(userClient *userclient.Client) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		token := r.GetHeader("Authorization")
		if token == "" {
			writeModelAuthError(r, "model api key is required")
			return
		}

		// Strip "Bearer " prefix if present.
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// Validate model API key via user-service.
		result, err := userClient.ValidateModelKey(r.GetCtx(), token)
		if err != nil {
			writeModelAuthError(r, "invalid model api key")
			return
		}

		// Check that the associated session is still active.
		_, err = userClient.IntrospectSession(r.GetCtx(), result.SessionId)
		if err != nil {
			writeModelAuthError(r, "associated session has expired or been revoked")
			return
		}

		// Store validated info in context for downstream use.
		r.SetCtxVar("modelApiKey", token)
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
