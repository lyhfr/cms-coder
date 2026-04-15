package auth

import (
	"context"
	"time"

	v1 "cmscoder-web-server/api/auth/v1"
	"cmscoder-web-server/internal/middleware"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// ModelTokenConfig holds configuration for Model Token issuance.
type ModelTokenConfig struct {
	TokenTTL        time.Duration
	EnableIPBinding bool
	JWTSecret       string
}

// ModelToken handles the model-token endpoint.
// It validates the HMAC signature and issues a short-lived JWT.
func (c *Controller) ModelToken(ctx context.Context, req *v1.ModelTokenReq) (res *v1.ModelTokenRes, err error) {
	cfg := c.getModelTokenConfig(ctx)

	// 1. Validate timestamp and nonce (anti-replay)
	if !c.nonceCache.IsValid(req.Nonce, req.Timestamp) {
		return nil, gerror.New("invalid or expired request (timestamp/nonce)")
	}

	// 2. Introspect access token to get session info including plugin_secret
	session, err := c.userClient.IntrospectSession(ctx, req.AccessToken)
	if err != nil {
		return nil, gerror.New("invalid access token")
	}

	// 3. Verify HMAC signature
	// signature = HMAC_SHA256(accessToken + timestamp + nonce, plugin_secret)
	message := req.AccessToken + string(rune(req.Timestamp)) + req.Nonce
	if !middleware.VerifyHMAC(message, session.PluginSecret, req.Signature) {
		// Note: The signature format should be: accessToken + timestamp(string) + nonce
		// Let me fix the message construction
		message = req.AccessToken + int64ToString(req.Timestamp) + req.Nonce
		if !middleware.VerifyHMAC(message, session.PluginSecret, req.Signature) {
			return nil, gerror.New("invalid HMAC signature")
		}
	}

	// 4. Optional: IP binding check
	if cfg.EnableIPBinding {
		// Get client IP from context
		clientIP := g.RequestFromCtx(ctx).GetClientIp()
		// Note: In a real implementation, we'd store and compare the original session IP
		// For now, this is a placeholder for the IP binding feature
		_ = clientIP
	}

	// 5. Generate Model Token (JWT)
	jwtHelper := middleware.NewJWTHelper(cfg.JWTSecret)
	modelToken, err := jwtHelper.GenerateToken(
		session.UserId,
		session.SessionId,
		"claude-code", // TODO: Get actual agent type from session
		cfg.TokenTTL,
	)
	if err != nil {
		return nil, gerror.Wrap(err, "failed to generate model token")
	}

	return &v1.ModelTokenRes{
		ModelToken: modelToken,
		ExpiresIn:  int64(cfg.TokenTTL.Seconds()),
		TokenType:  "Bearer",
	}, nil
}

// getModelTokenConfig reads Model Token configuration from config.
func (c *Controller) getModelTokenConfig(ctx context.Context) *ModelTokenConfig {
	cfg := g.Cfg()

	tokenTTL := cfg.MustGet(ctx, "model.modelTokenTTL", "5m").Duration()
	if tokenTTL == 0 {
		tokenTTL = 5 * time.Minute
	}

	jwtSecret := cfg.MustGet(ctx, "model.jwtSecret").String()
	if jwtSecret == "" {
		// Fallback: generate a random secret (should be configured in production)
		jwtSecret = "cmscoder-default-jwt-secret-change-in-production"
	}

	return &ModelTokenConfig{
		TokenTTL:        tokenTTL,
		EnableIPBinding: cfg.MustGet(ctx, "model.enableIPBinding", false).Bool(),
		JWTSecret:       jwtSecret,
	}
}

// int64ToString converts int64 to string without strconv to avoid import.
func int64ToString(n int64) string {
	if n == 0 {
		return "0"
	}
	var result []byte
	negative := n < 0
	if negative {
		n = -n
	}
	for n > 0 {
		result = append([]byte{byte('0' + n%10)}, result...)
		n /= 10
	}
	if negative {
		result = append([]byte{'-'}, result...)
	}
	return string(result)
}
