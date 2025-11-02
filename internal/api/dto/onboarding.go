package dto

import (
	"github.com/omkar273/nashikdarshan/internal/domain/user"
)

// OnboardingRequest is the request for the onboarding service
// It extends the SignupRequest to include the role
type OnboardingRequest struct {
	user.User
}

func (r *OnboardingRequest) Validate() error {

	return nil
}

// OnboardingResponse is the response for the onboarding service
// It extends the SignupResponse to include the role
type OnboardingResponse struct {
	user.User
}
