package userclient

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/gclient"
)

// Client is the HTTP client for communicating with user-service.
type Client struct {
	baseURL string
	httpCli *gclient.Client
}

// New creates a new user-service client.
func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpCli: g.Client(),
	}
}

// LoginSessionInput is the input for creating a login session.
type LoginSessionInput struct {
	LocalPort        int    `json:"localPort"`
	AgentType        string `json:"agentType"`
	PluginInstanceId string `json:"pluginInstanceId"`
	ClientVersion    string `json:"clientVersion"`
}

// LoginSessionOutput is the output of creating a login session.
type LoginSessionOutput struct {
	LoginId   string `json:"loginId"`
	BrowserUrl string `json:"browserUrl"`
	ExpiresAt string `json:"expiresAt"`
}

// CreateLoginSession calls user-service to create a login session.
func (c *Client) CreateLoginSession(ctx context.Context, in LoginSessionInput) (*LoginSessionOutput, error) {
	res, err := c.httpCli.Post(ctx, c.baseURL+"/user-service/auth/login", in)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	body := res.ReadAllString()
	var apiRes struct {
		Code int                `json:"code"`
		Data LoginSessionOutput `json:"data"`
	}
	if err := gjson.Unmarshal([]byte(body), &apiRes); err != nil {
		return nil, err
	}
	if apiRes.Code != 0 {
		return nil, gerror.Newf("user-service error, code: %d", apiRes.Code)
	}
	return &apiRes.Data, nil
}

// CallbackInput is the input for completing OAuth callback.
type CallbackInput struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// CallbackOutput is the output of completing OAuth callback.
type CallbackOutput struct {
	LoopbackRedirectUrl string `json:"loopbackRedirectUrl"`
}

// CompleteCallback calls user-service to complete OAuth callback.
func (c *Client) CompleteCallback(ctx context.Context, in CallbackInput) (*CallbackOutput, error) {
	res, err := c.httpCli.Post(ctx, c.baseURL+"/user-service/auth/iam/callback/complete", in)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	body := res.ReadAllString()
	var apiRes struct {
		Code int              `json:"code"`
		Data CallbackOutput   `json:"data"`
	}
	if err := gjson.Unmarshal([]byte(body), &apiRes); err != nil {
		return nil, err
	}
	if apiRes.Code != 0 {
		return nil, gerror.Newf("user-service error, code: %d", apiRes.Code)
	}
	return &apiRes.Data, nil
}

// ExchangeInput is the input for exchanging login ticket.
type ExchangeInput struct {
	LoginTicket      string `json:"loginTicket"`
	PluginInstanceId string `json:"pluginInstanceId"`
}

// ExchangeOutput is the output of exchanging login ticket.
type ExchangeOutput struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
	ModelApiKey  string `json:"modelApiKey"`
	User         User   `json:"user"`
}

// User contains user information.
type User struct {
	UserId      string `json:"userId"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	TenantId    string `json:"tenantId"`
}

// ExchangeLoginTicket calls user-service to exchange login ticket.
func (c *Client) ExchangeLoginTicket(ctx context.Context, in ExchangeInput) (*ExchangeOutput, error) {
	res, err := c.httpCli.Post(ctx, c.baseURL+"/user-service/auth/login-tickets/exchange", in)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	body := res.ReadAllString()
	var apiRes struct {
		Code int            `json:"code"`
		Data ExchangeOutput `json:"data"`
	}
	if err := gjson.Unmarshal([]byte(body), &apiRes); err != nil {
		return nil, err
	}
	if apiRes.Code != 0 {
		return nil, gerror.Newf("user-service error, code: %d", apiRes.Code)
	}
	return &apiRes.Data, nil
}

// RefreshInput is the input for refreshing session.
type RefreshInput struct {
	RefreshToken string `json:"refreshToken"`
}

// RefreshOutput is the output of refreshing session.
type RefreshOutput struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

// RefreshSession calls user-service to refresh session.
func (c *Client) RefreshSession(ctx context.Context, in RefreshInput) (*RefreshOutput, error) {
	res, err := c.httpCli.Post(ctx, c.baseURL+"/user-service/auth/sessions/refresh", in)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	body := res.ReadAllString()
	var apiRes struct {
		Code int           `json:"code"`
		Data RefreshOutput `json:"data"`
	}
	if err := gjson.Unmarshal([]byte(body), &apiRes); err != nil {
		return nil, err
	}
	if apiRes.Code != 0 {
		return nil, gerror.Newf("user-service error, code: %d", apiRes.Code)
	}
	return &apiRes.Data, nil
}

// RevokeSession calls user-service to revoke a session.
func (c *Client) RevokeSession(ctx context.Context, sessionId, refreshToken string) error {
	res, err := c.httpCli.Post(ctx, c.baseURL+"/user-service/auth/sessions/revoke", g.Map{
		"sessionId":    sessionId,
		"refreshToken": refreshToken,
	})
	if err != nil {
		return err
	}
	defer res.Close()

	body := res.ReadAllString()
	var apiRes struct {
		Code int `json:"code"`
	}
	if err := gjson.Unmarshal([]byte(body), &apiRes); err != nil {
		return err
	}
	if apiRes.Code != 0 {
		return gerror.Newf("user-service error, code: %d", apiRes.Code)
	}
	return nil
}

// IntrospectResult is the result of introspecting a session.
type IntrospectResult struct {
	UserId      string `json:"userId"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	TenantId    string `json:"tenantId"`
	SessionId   string `json:"sessionId"`
	ExpiresAt   string `json:"expiresAt"`
}

// IntrospectSession calls user-service to introspect a session.
func (c *Client) IntrospectSession(ctx context.Context, accessToken string) (*IntrospectResult, error) {
	res, err := c.httpCli.Get(ctx, c.baseURL+"/user-service/auth/sessions/introspect", g.Map{
		"accessToken": accessToken,
	})
	if err != nil {
		return nil, err
	}
	defer res.Close()

	body := res.ReadAllString()
	var apiRes struct {
		Code int              `json:"code"`
		Data IntrospectResult `json:"data"`
	}
	if err := gjson.Unmarshal([]byte(body), &apiRes); err != nil {
		return nil, err
	}
	if apiRes.Code != 0 {
		return nil, gerror.Newf("user-service error, code: %d", apiRes.Code)
	}
	return &apiRes.Data, nil
}

// ModelKeyInfo contains validated model key information.
type ModelKeyInfo struct {
	UserId         string `json:"userId"`
	SessionId      string `json:"sessionId"`
	AgentType      string `json:"agentType"`
	PluginInstance string `json:"pluginInstance"`
	ExpiresAt      string `json:"expiresAt"`
}

// ValidateModelKey calls user-service to validate a model API key.
func (c *Client) ValidateModelKey(ctx context.Context, modelApiKey string) (*ModelKeyInfo, error) {
	res, err := c.httpCli.Post(ctx, c.baseURL+"/user-service/auth/model-keys/validate", g.Map{
		"modelApiKey": modelApiKey,
	})
	if err != nil {
		return nil, err
	}
	defer res.Close()

	body := res.ReadAllString()
	var apiRes struct {
		Code int          `json:"code"`
		Data ModelKeyInfo `json:"data"`
	}
	if err := gjson.Unmarshal([]byte(body), &apiRes); err != nil {
		return nil, err
	}
	if apiRes.Code != 0 {
		return nil, gerror.Newf("user-service error, code: %d", apiRes.Code)
	}
	return &apiRes.Data, nil
}
