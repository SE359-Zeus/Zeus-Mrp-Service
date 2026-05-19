package service

import "errors"

var (
	ErrNotFound        = errors.New("resource not found")
	ErrDuplicateEmail  = errors.New("email already exists")
	ErrInactiveAccount = errors.New("account is inactive")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrInvalidInput    = errors.New("invalid input")
	ErrShortPassword   = errors.New("password must be at least 8 characters")
	ErrInvalidRole     = errors.New("invalid role: must be Admin, Editor, or Viewer")
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrEmptyEmail      = errors.New("email cannot be empty")
	ErrEmptyPassword   = errors.New("password cannot be empty")
	ErrEmptyName       = errors.New("full name cannot be empty")
	ErrNilID           = errors.New("id is required")
)
