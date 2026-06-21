package mappers

import (
	"bookstore-backend/internal/dto"
	"bookstore-backend/internal/models"
)

func ToUserResponse(user *models.User) dto.UserResponse {
	return dto.UserResponse{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		Role:        user.Role,
		CreatedAt:   user.CreatedAt,
	}
}
