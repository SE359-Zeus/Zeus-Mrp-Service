package controllers

import (
	"net/http"

	"zeus-sales-service/internal/models"
	"zeus-sales-service/internal/service"
)

type OrderController struct {
	svc *service.OrderService
}

func NewOrderController(svc *service.OrderService) *OrderController {
	return &OrderController{svc: svc}
}

func (controller *OrderController) HandleOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var req models.CreateOrderRequest
		if err := readJSON(r, &req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		order, err := controller.svc.CreateOrder(r.Context(), req)
		if err != nil {
			panic(err)
		}
		writeJSON(w, http.StatusCreated, order)
	case http.MethodGet:
		orders, err := controller.svc.ListOrders(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, orders)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (controller *OrderController) HandleOrderByID(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r.URL.Path, "/api/v1/sales/orders/")
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid order id"})
		return
	}
	switch r.Method {
	case http.MethodGet:
		order, err := controller.svc.GetOrder(r.Context(), id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, order)
	case http.MethodPatch:
		var req models.UpdateOrderRequest
		if err := readJSON(r, &req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		order, err := controller.svc.UpdateOrder(r.Context(), id, req)
		if err != nil {
			panic(err)
		}
		writeJSON(w, http.StatusOK, order)
	case http.MethodDelete:
		if err := controller.svc.CancelOrder(r.Context(), id); err != nil {
			panic(err)
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "order cancelled"})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
