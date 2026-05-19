package service

import (
	"context"
	"math"
	"time"

	"zeus-system-service/internal/models"
	"zeus-system-service/internal/repository"

	"github.com/google/uuid"
)

type auditService struct {
	repo repository.AuditRepository
}

func NewAuditService(repo repository.AuditRepository) AuditService {
	return &auditService{repo: repo}
}

func (s *auditService) Ingest(ctx context.Context, req models.IngestAuditRequest) error {
	if req.ActionType == "" {
		return ErrInvalidInput
	}
	if !models.ValidActionTypes[req.ActionType] {
		return ErrInvalidInput
	}
	if req.UserID == uuid.Nil {
		return ErrInvalidInput
	}
	if req.UserEmail == "" {
		return ErrInvalidInput
	}
	if req.TargetResource == "" {
		return ErrInvalidInput
	}

	log := &models.AuditLog{
		UserID:          req.UserID,
		UserEmail:       req.UserEmail,
		ActionType:      req.ActionType,
		TargetResource:  req.TargetResource,
		Details:         req.Details,
		IPAddress:       req.IPAddress,
		IsSecurityEvent: req.IsSecurityEvent || req.ActionType == models.ActionSecurity,
	}

	return s.repo.Insert(ctx, log)
}

func (s *auditService) Query(ctx context.Context, filter models.AuditFilter, page, limit int) ([]models.AuditLog, *models.PaginationMeta, error) {
	logs, totalRows, err := s.repo.Query(ctx, filter, page, limit)
	if err != nil {
		return nil, nil, err
	}
	if logs == nil {
		logs = []models.AuditLog{}
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 15
	}
	totalPages := int(math.Ceil(float64(totalRows) / float64(limit)))
	meta := &models.PaginationMeta{
		Page:       page,
		Limit:      limit,
		TotalRows:  totalRows,
		TotalPages: totalPages,
	}
	return logs, meta, nil
}

func (s *auditService) GetMetrics(ctx context.Context) (*models.AuditMetrics, error) {
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	logins, err := s.repo.CountByAction(ctx, models.ActionLogin, today, tomorrow)
	if err != nil {
		return nil, err
	}

	securityEvents, err := s.repo.CountByAction(ctx, models.ActionSecurity, today, tomorrow)
	if err != nil {
		return nil, err
	}

	updates, err := s.repo.CountByAction(ctx, models.ActionUpdate, today, tomorrow)
	if err != nil {
		return nil, err
	}

	deletes, err := s.repo.CountByAction(ctx, models.ActionDelete, today, tomorrow)
	if err != nil {
		return nil, err
	}

	return &models.AuditMetrics{
		LoginsToday:          logins,
		SecurityEvents:       securityEvents,
		ModificationVelocity: updates + deletes,
	}, nil
}
