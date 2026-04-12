package auth

import (
	"context"
	"time"

	v1 "cmscoder-user-service/api/auth/v1"
	"cmscoder-user-service/internal/service/modelkey"
)

// ModelKeyValidate validates a model API key.
func (c *Controller) ModelKeyValidate(ctx context.Context, req *v1.ModelKeyValidateReq) (res *v1.ModelKeyValidateRes, err error) {
	out, err := c.modelKeySvc.ValidateModelKey(ctx, modelkey.ValidateInput{
		ModelApiKey: req.ModelApiKey,
	})
	if err != nil {
		return nil, err
	}

	return &v1.ModelKeyValidateRes{
		UserId:         out.UserId,
		SessionId:      out.SessionId,
		AgentType:      out.AgentType,
		PluginInstance: out.PluginInstance,
		ExpiresAt:      out.ExpiresAt.Format(time.RFC3339),
	}, nil
}
