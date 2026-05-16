package controllers

import "net/http"

// POST /api/v1/production/orders
func (c *ProductionController) CreateOrder(w http.ResponseWriter, r *http.Request) {
	// Logic to decode request and call c.svc.PlanProduction
}
