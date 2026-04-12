package auth

import (
	"cmscoder-web-server/internal/clients/userclient"
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
}

// New creates a new auth controller.
func New(userClient *userclient.Client, iamCfg *IAMConfig) *Controller {
	return &Controller{
		userClient: userClient,
		iamCfg:     iamCfg,
	}
}

// getUserClient returns the user-service client.
func (c *Controller) getUserClient() *userclient.Client {
	if c.userClient == nil {
		panic(gerror.New("user-service client is not initialized"))
	}
	return c.userClient
}
