package controllers

import (
	"net/http"

	"zeus-sales-service/internal/models"
	"zeus-sales-service/internal/service"
)

type ClientController struct {
	svc *service.ClientService
}

func NewClientController(svc *service.ClientService) *ClientController {
	return &ClientController{svc: svc}
}

func (controller *ClientController) HandleClients(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		clients, err := controller.svc.ListClients(r.Context())
		if err != nil {
			panic(err)
		}
		writeJSON(w, http.StatusOK, clients)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (controller *ClientController) HandleClientByID(w http.ResponseWriter, r *http.Request) {
	clientID, ok := parseID(r.URL.Path, "/api/v1/sales/clients/")
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid client id"})
		return
	}
	switch r.Method {
	case http.MethodPatch:
		var req models.UpdateClientRequest
		if err := readJSON(r, &req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		client, err := controller.svc.UpdateClient(r.Context(), clientID, req)
		if err != nil {
			panic(err)
		}
		writeJSON(w, http.StatusOK, client)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
