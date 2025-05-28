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

	// Instanciar un servidor con gin
	server := httptest.NewServer(r)
	defer server.Close()

	// Inicializar las rutas, un recorder y 2 variables auxiliares
	api.InitRoutes(r, server.URL) // URL base para las llamadas entre servicios
	var createdUser user.User
	var createdSale sales.Sales
	userRecorder := httptest.NewRecorder()

	// 1. Crear un usuario para luego crear una venta para el mismo (POST /users y POST /sales)
	userData := map[string]string{
		"name":     "Juancito",
		"address":  "suyuque",
		"nickname": "simplementeJuancito",
	}
	userBody, _ := json.Marshal(userData)
	userReq, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(userBody))
	r.ServeHTTP(userRecorder, userReq)

	json.Unmarshal(userRecorder.Body.Bytes(), &createdUser)

	// Usuario creado y parseado a la variable createdUser, ahora creo venta
	saleData := map[string]interface{}{
		"user_id": createdUser.ID,
		"amount":  100.50,
	}
	saleBody, _ := json.Marshal(saleData)

	saleReq, _ := http.NewRequest(http.MethodPost, "/sales", bytes.NewBuffer(saleBody))

	saleRecorder := httptest.NewRecorder()
	r.ServeHTTP(saleRecorder, saleReq)

	// Una vez creada la venta, chequeo con asserts que me devuelve lo correcto.

	assert.Equal(t, http.StatusCreated, saleRecorder.Code)

	err := json.Unmarshal(saleRecorder.Body.Bytes(), &createdSale)
	assert.NoError(t, err)
	assert.NotEmpty(t, createdSale.ID)
	assert.Equal(t, createdUser.ID, createdSale.UserID)
	assert.Equal(t, float32(100.50), createdSale.Amount)
	// Esto sirve para los asserts del final del get, por si se crean varias ventas hasta conseguir pending
	quantity_sales := 1
	amount_sales := createdSale.Amount

	// 2. Actualizar el estado de la venta (PATCH /sales/:id)
	// Solo actualizar si la venta está en estado pending (puede ser aleatorio)
	// Para asegurar que entre siempre a patchear creamos ventas hasta que haya una pending.
	maxAttempts := 10 // Evitar bucle infinito
	attempts := 0

	for attempts < maxAttempts {
		if createdSale.Status == "pending" {
			updateData := map[string]string{
				"status": "approved",
			}
			updateBody, _ := json.Marshal(updateData)

			// Crear un nuevo recorder para esta petición
			patchRecorder := httptest.NewRecorder()
			patchReq, _ := http.NewRequest(http.MethodPatch, "/sales/"+createdSale.ID, bytes.NewBuffer(updateBody))
			r.ServeHTTP(patchRecorder, patchReq)

			assert.Equal(t, http.StatusOK, patchRecorder.Code)

			var updatedSale sales.Sales
			err = json.Unmarshal(patchRecorder.Body.Bytes(), &updatedSale)
			assert.NoError(t, err)
			assert.Equal(t, "approved", updatedSale.Status)
			break
		}

		// Crear nueva venta
		pendingRecorder := httptest.NewRecorder()
		pendingReq, _ := http.NewRequest(http.MethodPost, "/sales", bytes.NewBuffer(saleBody))
		r.ServeHTTP(pendingRecorder, pendingReq)

		err = json.Unmarshal(pendingRecorder.Body.Bytes(), &createdSale)
		assert.NoError(t, err)

		quantity_sales++
		amount_sales += createdSale.Amount
		attempts++
	}

	if attempts == maxAttempts {
		t.Fatal("No se pudo encontrar una venta con estado 'pending' después de múltiples intentos")
	}

	// 3. Obtener las ventas del usuario (GET /sales?user_id=...)
	req, _ := http.NewRequest(http.MethodGet, "/sales?user_id="+createdUser.ID, nil)
	getRecorder := httptest.NewRecorder()
	r.ServeHTTP(getRecorder, req)

	// Chequeamos con asserts que el get traiga lo que corresponde, utilizando las variables para contar que utilizamos arriba
	assert.Equal(t, http.StatusOK, getRecorder.Code)

	var response api.SalesResponse
	err = json.Unmarshal(getRecorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, quantity_sales, response.Metadata.Quantity)
	assert.InDelta(t, amount_sales, response.Metadata.TotalAmount, 0.1)
}
