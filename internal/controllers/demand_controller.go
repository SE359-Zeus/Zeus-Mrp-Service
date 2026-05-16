package controllers

import "net/http"

// GET /api/v1/mrp/demand
func (c *ProductionController) GetDemandSummary(w http.ResponseWriter, r *http.Request) {
	// TODO: Call c.svc.GetDemandSummary
}

// POST /api/v1/mrp/demand/generate-pos
func (c *ProductionController) GeneratePOs(w http.ResponseWriter, r *http.Request) {
	// TODO: Call c.svc.GeneratePOsForShortages
}

// GET /api/v1/mrp/demand/{orderId}/pick-list
func (c *ProductionController) GetPickList(w http.ResponseWriter, r *http.Request) {
	// TODO: Call c.svc.GeneratePickList
}
