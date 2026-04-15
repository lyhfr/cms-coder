package middleware

import (
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// ModelTokenAuth validates the JWT Model Token from the Authorization header.
// It extracts the JWT, verifies the signature and expiration, and stores claims in context.
func ModelTokenAuth() ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		token := r.GetHeader("Authorization")
		if token == "" {
			writeModelTokenAuthError(r, "authorization header is required")
			return
		}

		// Strip "Bearer " prefix if present.
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// Get JWT secret from config
		jwtSecret := g.Cfg().MustGet(r.GetCtx(), "model.jwtSecret").String()
		if jwtSecret == "" {
			jwtSecret = "cmscoder-default-jwt-secret-change-in-production"
		}

		// Verify JWT
		jwtHelper := NewJWTHelper(jwtSecret)
		claims, err := jwtHelper.VerifyToken(token)
		if err != nil {
			writeModelTokenAuthError(r, "invalid or expired model token")
			return
		}

		// Store validated info in context for downstream use.
		r.SetCtxVar("userId", claims.Subject)
		r.SetCtxVar("sessionId", claims.Session)
		r.SetCtxVar("agentType", claims.Agent)

		r.Middleware.Next()
	}
}

// writeModelAuthError writes a 401 unauthorized response for model endpoints.
func writeModelTokenAuthError(r *ghttp.Request, message string) {
	r.Response.WriteStatus(401)
	r.Response.WriteJson(ghttp.DefaultHandlerResponse{
		Code:    gcode.CodeNotAuthorized.Code(),
		Message: message,
	})
	r.Exit()
}
