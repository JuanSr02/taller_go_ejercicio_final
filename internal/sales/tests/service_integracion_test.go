package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ej_final/api"
	"ej_final/internal/sales"
	"ej_final/internal/user"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestService_Integracion_HappyPath(t *testing.T) {
	// Configurar Gin en modo test
	gin.SetMode(gin.TestMode)
	r := gin.New()

	server := httptest.NewServer(r)
	defer server.Close()

	api.InitRoutes(r, server.URL) // URL base para las llamadas entre servicios
	var createdUser user.User
	var createdSale sales.Sales
	w := httptest.NewRecorder()

	// 1. Crear una venta para el usuario (POST /sales)
	userData := map[string]string{
		"name":     "Juancito",
		"address":  "suyuque",
		"nickname": "simplementeJuancito",
	}
	userBody, _ := json.Marshal(userData)
	userReq, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(userBody))
	userReq.Header.Set("Content-Type", "application/json")
	userW := httptest.NewRecorder()
	r.ServeHTTP(userW, userReq)

	json.Unmarshal(userW.Body.Bytes(), &createdUser)

	saleData := map[string]interface{}{
		"user_id": createdUser.ID,
		"amount":  100.50,
	}
	saleBody, _ := json.Marshal(saleData)

	req, _ := http.NewRequest(http.MethodPost, "/sales", bytes.NewBuffer(saleBody))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	err := json.Unmarshal(w.Body.Bytes(), &createdSale)
	assert.NoError(t, err)
	assert.NotEmpty(t, createdSale.ID)
	assert.Equal(t, createdUser.ID, createdSale.UserID)
	assert.Equal(t, float32(100.50), createdSale.Amount)

	// 2. Actualizar el estado de la venta (PATCH /sales/:id)
	// Solo actualizar si la venta está en estado pending (puede ser aleatorio)
	if createdSale.Status == "pending" {
		updateData := map[string]string{
			"status": "approved",
		}
		updateBody, _ := json.Marshal(updateData)

		req, _ := http.NewRequest(http.MethodPatch, "/sales/"+createdSale.ID, bytes.NewBuffer(updateBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var updatedSale sales.Sales
		err := json.Unmarshal(w.Body.Bytes(), &updatedSale)
		assert.NoError(t, err)
		assert.Equal(t, "approved", updatedSale.Status)
	}

	// 3. Obtener las ventas del usuario (GET /sales?user_id=...)
	// Obtener ventas
	req, _ = http.NewRequest(http.MethodGet, "/sales?user_id="+createdUser.ID, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response api.SalesResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 1, response.Metadata.Quantity)
	assert.InDelta(t, 100.5, response.Metadata.TotalAmount, 0.1) // 50 + 100 + 150
}
