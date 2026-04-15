package auth

import (
	"cmscoder-web-server/internal/clients/userclient"
	"cmscoder-web-server/internal/middleware"

	"github.com/gogf/gf/v2/errors/gerror"
)

// IAMConfig holds IAM OAuth configuration for web-server.
type IAMConfig struct {
	AuthorizeURL string
	ClientID     string
	RedirectURI  string
}

// Controller is the auth API controller.
type Controller struct {
	userClient *userclient.Client
	iamCfg     *IAMConfig
	nonceCache *middleware.NonceCache
}

// New creates a new auth controller.
func New(userClient *userclient.Client, iamCfg *IAMConfig, nonceCache *middleware.NonceCache) *Controller {
	return &Controller{
		userClient: userClient,
		iamCfg:     iamCfg,
		nonceCache: nonceCache,
	}
}

// getUserClient returns the user-service client.
func (c *Controller) getUserClient() *userclient.Client {
	if c.userClient == nil {
		panic(gerror.New("user-service client is not initialized"))
	}
	return c.userClient
}
