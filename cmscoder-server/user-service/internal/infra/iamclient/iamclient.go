package iamclient

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/gclient"
)

// Client is the HTTP client for communicating with IAM.
type Client struct {
	tokenURL     string
	userInfoURL  string
	clientID     string
	clientSecret string
	httpCli      *gclient.Client
}

// New creates a new IAM client.
func New(tokenURL, userInfoURL, clientID, clientSecret string) *Client {
	return &Client{
		tokenURL:     tokenURL,
		userInfoURL:  userInfoURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		httpCli:      g.Client(),
	}
}

// TokenResponse is the response from IAM token endpoint.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// GetToken exchanges authorization code for access token.
func (c *Client) GetToken(ctx context.Context, code string) (*TokenResponse, error) {
	res, err := c.httpCli.Post(ctx, c.tokenURL, g.Map{
		"grant_type":    "authorization_code",
		"code":          code,
		"client_id":     c.clientID,
		"client_secret": c.clientSecret,
	})
	if err != nil {
		return nil, gerror.Wrap(err, "IAM GetToken request failed")
	}
	defer res.Close()

	body := res.ReadAllString()
	var tokenResp TokenResponse
	if err := gjson.Unmarshal([]byte(body), &tokenResp); err != nil {
		return nil, gerror.Wrapf(err, "failed to parse IAM token response: %s", body)
	}
	if tokenResp.AccessToken == "" {
		return nil, gerror.Newf("IAM returned empty access token: %s", body)
	}
	return &tokenResp, nil
}

// UserInfo is the response from IAM user info endpoint.
type UserInfo struct {
	IamUserId   string `json:"iamUserId"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	TenantId    string `json:"tenantId"`
}

// GetUserInfo retrieves user info from IAM.
func (c *Client) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	res, err := c.httpCli.Header(map[string]string{"Authorization": "Bearer " + accessToken}).Get(ctx, c.userInfoURL)
	if err != nil {
		return nil, gerror.Wrap(err, "IAM GetUserInfo request failed")
	}
	defer res.Close()

	body := res.ReadAllString()
	var userInfo UserInfo
	if err := gjson.Unmarshal([]byte(body), &userInfo); err != nil {
		return nil, gerror.Wrapf(err, "failed to parse IAM user info response: %s", body)
	}
	if userInfo.IamUserId == "" {
		return nil, gerror.Newf("IAM returned empty user ID: %s", body)
	}
	return &userInfo, nil
}
