package service

import (
	"zeus-mrp-service/internal/repository"
)

type ProductionService struct {
	repo repository.MRPRepository
}

func NewProductionService(repo repository.MRPRepository) *ProductionService {
	return &ProductionService{repo: repo}
}
