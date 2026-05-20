package service

import "errors"

var (
	ErrNotImplemented      = errors.New("not implemented")
	ErrLockExpired         = errors.New("lock has expired")
	ErrStateRegression     = errors.New("state regression is not allowed")
	ErrAlreadyLocked       = errors.New("resource already locked by another operator")
	ErrMonoVendorViolation = errors.New("purchase order must involve a single vendor")
	ErrInsufficientDeficit = errors.New("insufficient deficit in pool for this SKU")
	ErrNoOptimalSupplier   = errors.New("no optimal supplier found for the given SKU")
	ErrAgingQuarantine     = errors.New("component exceeds aging threshold")
	ErrNotFound            = errors.New("resource not found")
	ErrInvalidTransition   = errors.New("invalid state transition")
	ErrUnauthorized        = errors.New("unauthorized")
)
