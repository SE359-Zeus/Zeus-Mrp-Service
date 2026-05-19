package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"zeus-system-service/internal/handler"
	"zeus-system-service/internal/models"
	"zeus-system-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupAuthTest() (*gin.Engine, *handler.MockAuthService) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(handler.MockAuthService)
	h := handler.NewAuthHandler(mockSvc)
	r := gin.New()

	auth := r.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.Refresh)
	}

	return r, mockSvc
}

func TestAuthHandler_Login_200(t *testing.T) {
	r, mockSvc := setupAuthTest()

	req := models.LoginRequest{Email: "admin@zeus.com", Password: "pass123"}
	pair := &models.TokenPair{
		AccessToken:  "access-token-value",
		RefreshToken: "refresh-token-value",
		TokenType:    "Bearer",
		ExpiresIn:    900,
	}
	body, _ := json.Marshal(req)

	mockSvc.On("Login", mock.Anything, mock.AnythingOfType("models.LoginRequest")).Return(pair, nil)

	w := httptest.NewRecorder()
	reqHTTP, _ := http.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	reqHTTP.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, reqHTTP)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp struct {
		AccessToken string `json:"access_token"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, pair.AccessToken, resp.AccessToken)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Login_400_InvalidBody(t *testing.T) {
	r, _ := setupAuthTest()

	w := httptest.NewRecorder()
	reqHTTP, _ := http.NewRequest("POST", "/auth/login", bytes.NewReader([]byte(`not json`)))
	reqHTTP.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, reqHTTP)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthHandler_Login_401_InvalidCredentials(t *testing.T) {
	r, mockSvc := setupAuthTest()

	req := models.LoginRequest{Email: "admin@zeus.com", Password: "wrong"}
	body, _ := json.Marshal(req)

	mockSvc.On("Login", mock.Anything, mock.AnythingOfType("models.LoginRequest")).Return(nil, service.ErrUnauthorized)

	w := httptest.NewRecorder()
	reqHTTP, _ := http.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	reqHTTP.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, reqHTTP)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Login_401_Inactive(t *testing.T) {
	r, mockSvc := setupAuthTest()

	req := models.LoginRequest{Email: "inactive@zeus.com", Password: "pass"}
	body, _ := json.Marshal(req)

	mockSvc.On("Login", mock.Anything, mock.AnythingOfType("models.LoginRequest")).Return(nil, service.ErrInactiveAccount)

	w := httptest.NewRecorder()
	reqHTTP, _ := http.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	reqHTTP.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, reqHTTP)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Refresh_200(t *testing.T) {
	r, mockSvc := setupAuthTest()

	req := models.RefreshRequest{RefreshToken: "valid-refresh-token"}
	pair := &models.TokenPair{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
		TokenType:    "Bearer",
		ExpiresIn:    900,
	}
	body, _ := json.Marshal(req)

	mockSvc.On("Refresh", mock.Anything, mock.AnythingOfType("models.RefreshRequest")).Return(pair, nil)

	w := httptest.NewRecorder()
	reqHTTP, _ := http.NewRequest("POST", "/auth/refresh", bytes.NewReader(body))
	reqHTTP.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, reqHTTP)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.TokenPair
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Refresh_401(t *testing.T) {
	r, mockSvc := setupAuthTest()

	req := models.RefreshRequest{RefreshToken: "expired-or-invalid"}
	body, _ := json.Marshal(req)

	mockSvc.On("Refresh", mock.Anything, mock.AnythingOfType("models.RefreshRequest")).Return(nil, service.ErrUnauthorized)

	w := httptest.NewRecorder()
	reqHTTP, _ := http.NewRequest("POST", "/auth/refresh", bytes.NewReader(body))
	reqHTTP.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, reqHTTP)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Refresh_400(t *testing.T) {
	r, _ := setupAuthTest()

	w := httptest.NewRecorder()
	reqHTTP, _ := http.NewRequest("POST", "/auth/refresh", bytes.NewReader([]byte(`not json`)))
	reqHTTP.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, reqHTTP)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
