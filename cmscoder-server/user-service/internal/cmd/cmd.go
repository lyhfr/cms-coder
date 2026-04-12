package cmd

import (
	"context"

	"cmscoder-user-service/internal/cache"
	"cmscoder-user-service/internal/controller/auth"
	"cmscoder-user-service/internal/infra/iamclient"
	"cmscoder-user-service/internal/service/iamcallback"
	"cmscoder-user-service/internal/service/loginsession"
	"cmscoder-user-service/internal/service/modelkey"
	"cmscoder-user-service/internal/service/session"
	"cmscoder-user-service/internal/service/ticket"
	"cmscoder-user-service/internal/service/userprofile"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/goai"
	"github.com/gogf/gf/v2/os/gcmd"
)

const (
	OpenAPITitle       = `cmscoder User Service`
	OpenAPIDescription = `cmscoder user-service: SSO, user identity and session management core service`
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start cmscoder user-service HTTP server",
		Func:  mainFunc,
	}
)

func mainFunc(ctx context.Context, parser *gcmd.Parser) (err error) {
	var (
		s   = g.Server()
		cfg = g.Cfg()
	)

	// Initialize infrastructure.
	memCache := cache.New()

	iamCfg := struct {
		TokenURL     string
		UserInfoURL  string
		ClientID     string
		ClientSecret string
	}{
		TokenURL:     cfg.MustGet(ctx, "iam.tokenURL").String(),
		UserInfoURL:  cfg.MustGet(ctx, "iam.userInfoURL").String(),
		ClientID:     cfg.MustGet(ctx, "iam.clientId").String(),
		ClientSecret: cfg.MustGet(ctx, "iam.clientSecret").String(),
	}
	iamClient := iamclient.New(iamCfg.TokenURL, iamCfg.UserInfoURL, iamCfg.ClientID, iamCfg.ClientSecret)

	// Initialize services.
	loginSessionSvc := loginsession.New(memCache)
	ticketSvc := ticket.New(memCache)
	modelKeySvc := modelkey.New(memCache)
	sessionSvc := session.New(memCache, modelKeySvc)
	userProfileSvc := userprofile.New(memCache)
	iamCallbackSvc := iamcallback.New(
		iamClient,
		loginSessionSvc,
		ticketSvc,
		userProfileSvc,
		memCache,
		cfg.MustGet(ctx, "iam.redirectURI").String(),
	)

	// Initialize controller.
	authCtrl := auth.New(
		loginSessionSvc,
		iamCallbackSvc,
		ticketSvc,
		sessionSvc,
		userProfileSvc,
		modelKeySvc,
	)

	// Register routes.
	s.Use(ghttp.MiddlewareHandlerResponse)
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(ghttp.MiddlewareCORS)
		group.Bind(authCtrl)
	})

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
