package auth

import (
	"context"

	v1 "cmscoder-web-server/api/auth/v1"
	"cmscoder-web-server/internal/middleware"
	"github.com/gogf/gf/v2/net/ghttp"
)

// PluginBootstrap returns plugin bootstrap info.
func (c *Controller) PluginBootstrap(ctx context.Context, req *v1.PluginBootstrapReq) (res *v1.PluginBootstrapRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	token := middleware.ParseBearerToken(r)

	bootstrapRes := &v1.PluginBootstrapRes{
		FeatureFlags: map[string]bool{},
		DefaultModel: "",
		Status: v1.BootstrapStatus{
			IsLoggedIn: false,
		},
	}

	if token != "" {
		out, err := c.getUserClient().IntrospectSession(ctx, token)
		if err == nil {
			bootstrapRes.User = v1.UserInfo{
				UserId:      out.UserId,
				Email:       out.Email,
				DisplayName: out.DisplayName,
				TenantId:    out.TenantId,
			}
			bootstrapRes.Status.IsLoggedIn = true
		}
	}

	return bootstrapRes, nil
}
