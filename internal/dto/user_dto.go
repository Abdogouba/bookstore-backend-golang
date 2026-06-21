package dto

import "time"

type UpdateProfileRequest struct {
	Name string `json:"name" binding:"required,min=2,max=100"`

	Email string `json:"email" binding:"required,email,max=255"`

	PhoneNumber string `json:"phone_number" binding:"max=20"`
}

// UserResponse is returned to clients.
type UserResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phone_number"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
}

// ChangePasswordRequest contains the data needed
// to change the authenticated user's password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required,min=6,max=100"`

	NewPassword string `json:"new_password" binding:"required,min=6,max=100"`
}

// DeleteProfileRequest is required to confirm account deletion.
type DeleteProfileRequest struct {
	Password string `json:"password" binding:"required,min=6,max=100"`
}