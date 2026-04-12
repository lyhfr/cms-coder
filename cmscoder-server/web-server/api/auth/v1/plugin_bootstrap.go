package v1

import "github.com/gogf/gf/v2/frame/g"

// PluginBootstrapReq is the request for plugin bootstrap info.
type PluginBootstrapReq struct {
	g.Meta `path:"/api/plugin/bootstrap" method:"get" tags:"PluginService" summary:"Get plugin bootstrap info"`
}

// PluginBootstrapRes is the response for plugin bootstrap info.
type PluginBootstrapRes struct {
	User            UserInfo             `json:"user"`
	FeatureFlags    map[string]bool      `json:"featureFlags"`
	DefaultModel    string               `json:"defaultModel"`
	Status          BootstrapStatus      `json:"status"`
}

// BootstrapStatus contains basic status information.
type BootstrapStatus struct {
	IsLoggedIn bool `json:"isLoggedIn"`
}
