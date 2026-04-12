package auth

import (
	"context"
	"fmt"

	v1 "cmscoder-web-server/api/auth/v1"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// LoginSessionAuthorize redirects browser to IAM authorize page.
func (c *Controller) LoginSessionAuthorize(ctx context.Context, req *v1.LoginSessionAuthorizeReq) (res *v1.LoginSessionAuthorizeRes, err error) {
	if c.iamCfg == nil || c.iamCfg.AuthorizeURL == "" {
		return nil, gerror.New("IAM authorize URL is not configured")
	}

	authorizeURL := fmt.Sprintf("%s?response_type=code&client_id=%s&redirect_uri=%s&state=%s",
		c.iamCfg.AuthorizeURL,
		c.iamCfg.ClientID,
		c.iamCfg.RedirectURI,
		req.LoginId,
	)

	// Redirect browser to IAM authorize page.
	ghttp.RequestFromCtx(ctx).Response.RedirectTo(authorizeURL)
	return &v1.LoginSessionAuthorizeRes{}, nil
}
