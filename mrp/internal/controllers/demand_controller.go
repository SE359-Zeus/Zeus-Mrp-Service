package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

type demandPOGenerateResponse struct {
	Message string `json:"message"`
}

type demandErrorResponse struct {
	Error string `json:"error"`
}

func writeDemandJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, `{"error":"failed to encode response"}`, http.StatusInternalServerError)
	}
}

func writeDemandError(w http.ResponseWriter, statusCode int, message string) {
	writeDemandJSON(w, statusCode, demandErrorResponse{Error: message})
}

func parseDemandOrderID(r *http.Request) (uuid.UUID, error) {
	rawOrderID := r.PathValue("orderId")
	if rawOrderID == "" {
		return uuid.Nil, errors.New("orderId path parameter is required")
	}

	orderID, err := uuid.Parse(rawOrderID)
	if err != nil {
		return uuid.Nil, errors.New("orderId must be a valid UUID")
	}

	return orderID, nil
}

// GET /api/v1/mrp/demand
func (c *ProductionController) GetDemandSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := c.svc.GetDemandSummary(r.Context())
	if err != nil {
		writeDemandError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeDemandJSON(w, http.StatusOK, summary)
}

// POST /api/v1/mrp/demand/generate-pos
func (c *ProductionController) GeneratePOs(w http.ResponseWriter, r *http.Request) {
	if err := c.svc.GeneratePOsForShortages(r.Context()); err != nil {
		writeDemandError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeDemandJSON(w, http.StatusOK, demandPOGenerateResponse{Message: "PO generation started"})
}

// GET /api/v1/mrp/demand/{orderId}/pick-list
func (c *ProductionController) GetPickList(w http.ResponseWriter, r *http.Request) {
	orderID, err := parseDemandOrderID(r)
	if err != nil {
		writeDemandError(w, http.StatusBadRequest, err.Error())
		return
	}

	pickList, err := c.svc.GeneratePickList(r.Context(), orderID)
	if err != nil {
		writeDemandError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if pickList == nil {
		writeDemandError(w, http.StatusNotFound, "pick list not found")
		return
	}

	writeDemandJSON(w, http.StatusOK, pickList)
}

// POST /api/v1/mrp/demand/{orderId}/pick-list
func (c *ProductionController) GeneratePickList(w http.ResponseWriter, r *http.Request) {
	orderID, err := parseDemandOrderID(r)
	if err != nil {
		writeDemandError(w, http.StatusBadRequest, err.Error())
		return
	}

	pickList, err := c.svc.GeneratePickList(r.Context(), orderID)
	if err != nil {
		writeDemandError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if pickList == nil {
		writeDemandError(w, http.StatusNotFound, "pick list not found")
		return
	}

	writeDemandJSON(w, http.StatusCreated, pickList)
}
