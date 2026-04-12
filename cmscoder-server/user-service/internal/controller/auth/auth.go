package auth

import (
	"cmscoder-user-service/internal/service/iamcallback"
	"cmscoder-user-service/internal/service/loginsession"
	"cmscoder-user-service/internal/service/session"
	"cmscoder-user-service/internal/service/ticket"
	"cmscoder-user-service/internal/service/userprofile"
)

// Controller is the internal auth API controller.
type Controller struct {
	loginSessionSvc  *loginsession.Service
	iamCallbackSvc   *iamcallback.Service
	ticketSvc        *ticket.Service
	sessionSvc       *session.Service
	userProfileSvc   *userprofile.Service
}

// New creates a new auth controller.
func New(
	loginSessionSvc *loginsession.Service,
	iamCallbackSvc *iamcallback.Service,
	ticketSvc *ticket.Service,
	sessionSvc *session.Service,
	userProfileSvc *userprofile.Service,
) *Controller {
	return &Controller{
		loginSessionSvc: loginSessionSvc,
		iamCallbackSvc:  iamCallbackSvc,
		ticketSvc:       ticketSvc,
		sessionSvc:      sessionSvc,
		userProfileSvc:  userProfileSvc,
	}
}
