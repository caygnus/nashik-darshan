package dto

import "github.com/omkar273/codegeeky/internal/domain/user"

type MeResponse struct {
	*user.User
}

type UpdateUserRequest struct {
	Name  string `json:"name,omitempty"`
	Phone string `json:"phone,omitempty"`
}
