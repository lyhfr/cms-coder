package iamcallback

import (
	"context"
	"fmt"
	"time"

	"cmscoder-user-service/internal/cache"
	"cmscoder-user-service/internal/infra/iamclient"
	"cmscoder-user-service/internal/service/loginsession"
	"cmscoder-user-service/internal/service/ticket"
	"cmscoder-user-service/internal/service/userprofile"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/guid"
)

// Service handles IAM OAuth callback processing.
type Service struct {
	iamClient       *iamclient.Client
	loginSessionSvc *loginsession.Service
	ticketSvc       *ticket.Service
	userProfileSvc  *userprofile.Service
	cache           *cache.MemoryCache
	redirectBaseURL string
}

// New creates a new IAM callback service.
func New(
	iamClient *iamclient.Client,
	loginSessionSvc *loginsession.Service,
	ticketSvc *ticket.Service,
	userProfileSvc *userprofile.Service,
	cache *cache.MemoryCache,
	redirectBaseURL string,
) *Service {
	return &Service{
		iamClient:       iamClient,
		loginSessionSvc: loginSessionSvc,
		ticketSvc:       ticketSvc,
		userProfileSvc:  userProfileSvc,
		cache:           cache,
		redirectBaseURL: redirectBaseURL,
	}
}

// CompleteInput is the input for completing IAM callback.
type CompleteInput struct {
	Code  string
	State string
}

// CompleteOutput is the output for completing IAM callback.
type CompleteOutput struct {
	LoopbackRedirectUrl string
}

// Complete processes IAM OAuth callback: exchanges code for token, gets user info, creates session and login ticket.
func (s *Service) Complete(ctx context.Context, in CompleteInput) (*CompleteOutput, error) {
	// 1. Find login session by state.
	session, err := s.cache.GetLoginSessionByState(ctx, in.State)
	if err != nil {
		return nil, gerror.Wrap(err, "failed to find login session")
	}
	if session == nil {
		return nil, gerror.New("login session not found or state mismatch")
	}

	// 2. Exchange code for IAM token.
	tokenResp, err := s.iamClient.GetToken(ctx, in.Code)
	if err != nil {
		return nil, gerror.Wrap(err, "failed to exchange code for IAM token")
	}

	// 3. Get user info from IAM.
	userInfo, err := s.iamClient.GetUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, gerror.Wrap(err, "failed to get user info from IAM")
	}

	// 4. Upsert user.
	user, err := s.userProfileSvc.Upsert(ctx, userprofile.UpsertInput{
		IamUserId:   userInfo.IamUserId,
		Email:       userInfo.Email,
		DisplayName: userInfo.DisplayName,
		TenantId:    userInfo.TenantId,
	})
	if err != nil {
		return nil, gerror.Wrap(err, "failed to upsert user")
	}

	// 5. Create user session.
	sessionId := guid.S()
	refreshToken := guid.S()
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	if err := s.cache.CreateUserSession(ctx, &cache.UserSession{
		SessionId:      sessionId,
		UserId:         user.UserId,
		AgentType:      session.AgentType,
		PluginInstance: session.PluginInstanceId,
		RefreshToken:   refreshToken,
		ExpiresAt:      expiresAt,
	}, 7*24*time.Hour); err != nil {
		return nil, gerror.Wrap(err, "failed to create user session")
	}
	// Index user session by login ID for ticket exchange.
	if err := s.cache.SetUserSessionByLoginId(ctx, session.LoginId, sessionId, 7*24*time.Hour); err != nil {
		return nil, gerror.Wrap(err, "failed to index user session by login ID")
	}

	// 6. Generate login ticket.
	loginTicket := guid.S()
	if err := s.ticketSvc.Create(ctx, loginTicket, session.LoginId, session.PluginInstanceId, 60*time.Second); err != nil {
		return nil, gerror.Wrap(err, "failed to create login ticket")
	}

	// 7. Mark login session as completed.
	if err := s.cache.UpdateLoginSessionStatus(ctx, session.LoginId, "completed"); err != nil {
		return nil, gerror.Wrap(err, "failed to update login session status")
	}

	// 8. Build loopback redirect URL.
	loopbackRedirectUrl := fmt.Sprintf("http://127.0.0.1:%d/callback?login_ticket=%s", session.LocalPort, loginTicket)

	return &CompleteOutput{
		LoopbackRedirectUrl: loopbackRedirectUrl,
	}, nil
}
