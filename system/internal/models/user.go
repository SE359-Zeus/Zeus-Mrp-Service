package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AccountStatus string

const (
	AccountStatusActive   AccountStatus = "ACTIVE"
	AccountStatusInactive AccountStatus = "INACTIVE"
)

type UserRole string

const (
	UserRoleAdmin  UserRole = "Admin"
	UserRoleEditor UserRole = "Editor"
	UserRoleViewer UserRole = "Viewer"
)

var ValidRoles = map[UserRole]bool{
	UserRoleAdmin:  true,
	UserRoleEditor: true,
	UserRoleViewer: true,
}

type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Email        string         `gorm:"uniqueIndex;not null;size:255" json:"email"`
	PasswordHash string         `gorm:"not null" json:"-"`
	FullName     string         `gorm:"not null;size:255" json:"full_name"`
	Role         UserRole       `gorm:"not null;default:Viewer;size:20" json:"role"`
	Status       AccountStatus  `gorm:"not null;default:ACTIVE;size:20" json:"status"`
	LastLoginAt  *time.Time     `json:"last_login_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

type CreateUserRequest struct {
	Email    string   `json:"email" binding:"required"`
	Password string   `json:"password" binding:"required"`
	FullName string   `json:"full_name" binding:"required"`
	Role     UserRole `json:"role" binding:"required"`
}

type UpdateUserRequest struct {
	FullName *string   `json:"full_name,omitempty"`
	Role     *UserRole `json:"role,omitempty"`
}

type UserResponse struct {
	ID          uuid.UUID     `json:"id"`
	Email       string        `json:"email"`
	FullName    string        `json:"full_name"`
	Role        UserRole      `json:"role"`
	Status      AccountStatus `json:"status"`
	LastLoginAt *time.Time    `json:"last_login_at,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

func ToUserResponse(u *User) UserResponse {
	return UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		FullName:    u.FullName,
		Role:        u.Role,
		Status:      u.Status,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}
