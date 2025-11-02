package dto

import (
	"context"

	"github.com/omkar273/codegeeky/internal/domain/user"
	"github.com/omkar273/codegeeky/internal/types"
	"github.com/omkar273/codegeeky/internal/validator"
)

type SignupRequest struct {
	// basic info
	Email string `json:"email" validate:"email,required"`
	Phone string `json:"phone"`
	Name  string `json:"name" validate:"required"`

	// access token
	AccessToken string `json:"access_token" validate:"required"`
}

func (r *SignupRequest) ToUser(ctx context.Context) *user.User {
	return &user.User{
		ID:        types.GenerateUUIDWithPrefix(types.UUID_PREFIX_USER),
		Email:     r.Email,
		Phone:     r.Phone,
		Name:      r.Name,
		Role:      types.UserRoleUser,
		BaseModel: types.GetDefaultBaseModel(ctx),
	}
}

func (r *SignupRequest) Validate() error {
	if err := validator.ValidateRequest(r); err != nil {
		return err
	}

	return nil
}

type SignupResponse struct {
	ID          string `json:"id"`
	AccessToken string `json:"access_token"`
}
