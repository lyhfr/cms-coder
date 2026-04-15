package cmd

import (
	"context"
	"time"

	"cmscoder-web-server/internal/clients/userclient"
	"cmscoder-web-server/internal/controller/auth"
	"cmscoder-web-server/internal/controller/model"
	"cmscoder-web-server/internal/middleware"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/goai"
	"github.com/gogf/gf/v2/os/gcmd"
)

const (
	OpenAPITitle       = `cmscoder Web Server`
	OpenAPIDescription = `cmscoder web-server: unified auth entry point and external routing layer`
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start cmscoder web-server HTTP server",
		Func:  mainFunc,
	}
)

func mainFunc(ctx context.Context, parser *gcmd.Parser) (err error) {
	userServiceBaseURL := g.Cfg().MustGet(ctx, "userService.baseURL").String()
	if userServiceBaseURL == "" {
		userServiceBaseURL = "http://127.0.0.1:39011"
	}

	iamCfg := &auth.IAMConfig{
		AuthorizeURL: g.Cfg().MustGet(ctx, "iam.authorizeURL").String(),
		ClientID:     g.Cfg().MustGet(ctx, "iam.clientId").String(),
		RedirectURI:  g.Cfg().MustGet(ctx, "iam.redirectURI").String(),
	}

	var (
		s               = g.Server()
		userClient      = userclient.New(userServiceBaseURL)
		nonceCache      = middleware.NewNonceCache(5 * time.Minute)
		authCtrl        = auth.New(userClient, iamCfg, nonceCache)
		rateLimiter     = middleware.NewRateLimiter(100, time.Minute)
		upstreamBaseURL = g.Cfg().MustGet(ctx, "model.upstreamBaseURL").String()
		upstreamApiKey  = g.Cfg().MustGet(ctx, "model.upstreamApiKey").String()
		defaultModel    = g.Cfg().MustGet(ctx, "model.defaultModel").String()
		modelCtrl       = model.New(upstreamBaseURL, upstreamApiKey, defaultModel)
	)

	s.Use(ghttp.MiddlewareHandlerResponse)
	s.Group("/", func(group *ghttp.RouterGroup) {
		// Global middlewares.
		group.Middleware(
			middleware.Tracing,
			rateLimiter.Middleware,
			ghttp.MiddlewareCORS,
		)

		// Public auth endpoints (no authentication required).
		group.Bind(authCtrl)

		// Model endpoints with JWT Model Token auth.
		group.Group("/", func(group *ghttp.RouterGroup) {
			group.Middleware(middleware.ModelTokenAuth())
			group.Bind(modelCtrl)
		})

		// Protected endpoints requiring authentication.
		group.Group("/", func(group *ghttp.RouterGroup) {
			group.Middleware(middleware.Auth(userClient))
			group.ALLMap(g.Map{
				"/api/auth/me":          authCtrl.Me,
				"/api/plugin/bootstrap": authCtrl.PluginBootstrap,
			})
		})
	})

	// Customize OpenAPI documentation.
	enhanceOpenAPIDoc(s)

	s.Run()
	return nil
}

func enhanceOpenAPIDoc(s *ghttp.Server) {
	openapi := s.GetOpenApi()
	openapi.Config.CommonResponse = ghttp.DefaultHandlerResponse{}
	openapi.Config.CommonResponseDataField = `Data`

	openapi.Info = goai.Info{
		Title:       OpenAPITitle,
		Description: OpenAPIDescription,
		Contact: &goai.Contact{
			Name: "cmscoder",
		},
	}
}
