package controllers

import (
	"zeus-mrp-service/internal/service"
)

type ProductionController struct {
	svc *service.ProductionService
}

func NewProductionController(svc *service.ProductionService) *ProductionController {
	return &ProductionController{svc: svc}
}
