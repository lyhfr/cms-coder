package middleware

import (
	"cmscoder-web-server/internal/clients/userclient"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// Auth verifies the access token from the Authorization header.
func Auth(userClient *userclient.Client) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		token := r.GetHeader("Authorization")
		if token == "" {
			r.Response.WriteStatus(401)
			r.Response.WriteJson(ghttp.DefaultHandlerResponse{
				Code:    gcode.CodeNotAuthorized.Code(),
				Message: "access token is required",
			})
			return
		}

		// Strip "Bearer " prefix if present.
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		_, err := userClient.IntrospectSession(r.GetCtx(), token)
		if err != nil {
			r.Response.WriteStatus(401)
			r.Response.WriteJson(ghttp.DefaultHandlerResponse{
				Code:    gcode.CodeNotAuthorized.Code(),
				Message: "invalid or expired access token",
			})
			return
		}

		r.SetCtxVar("accessToken", token)
		r.Middleware.Next()
	}
}

// ParseBearerToken extracts the bearer token from request header for endpoints
// that accept token optionally.
func ParseBearerToken(r *ghttp.Request) string {
	token := r.GetHeader("Authorization")
	if token == "" {
		return ""
	}
	if len(token) > 7 && token[:7] == "Bearer " {
		return token[7:]
	}
	return token
}

// AuthError creates a standard unauthorized error response.
func AuthError(r *ghttp.Request, message string) {
	r.Response.WriteStatus(401)
	r.Response.WriteJson(ghttp.DefaultHandlerResponse{
		Code:    gcode.CodeNotAuthorized.Code(),
		Message: message,
	})
	r.Exit()
}

// ValidationFailed creates a standard bad request error response.
func ValidationFailed(r *ghttp.Request, message string) {
	r.Response.WriteStatus(400)
	r.Response.WriteJson(ghttp.DefaultHandlerResponse{
		Code:    gcode.CodeValidationFailed.Code(),
		Message: message,
	})
	r.Exit()
}

// NewValidationError creates a validation error using gerror.
func NewValidationError(message string) error {
	return gerror.NewCode(gcode.New(400, message, nil), message)
}

// NewAuthError creates an authentication error using gerror.
func NewAuthError(message string) error {
	return gerror.NewCode(gcode.New(401, message, nil), message)
}
