package api

import (
	"ej_final/internal/sales"
	"ej_final/internal/user"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type handler struct {
	userService  *user.Service
	salesService *sales.Service
	logger       *zap.Logger
}

// handleCreate handles POST /users
func (h *handler) handleCreate(ctx *gin.Context) {
	// request payload
	var req struct {
		Name     string `json:"name"`
		Address  string `json:"address"`
		NickName string `json:"nickname"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u := &user.User{
		Name:     req.Name,
		Address:  req.Address,
		NickName: req.NickName,
	}
	if err := h.userService.Create(u); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("user created", zap.Any("user", u))
	ctx.JSON(http.StatusCreated, u)
}

// handleRead handles GET /users/:id
func (h *handler) handleRead(ctx *gin.Context) {
	id := ctx.Param("id")

	u, err := h.userService.Get(id)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			h.logger.Warn("user not found", zap.String("id", id))
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		h.logger.Error("error trying to get user", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("get user succeed", zap.Any("user", u))
	ctx.JSON(http.StatusOK, u)
}

// handleUpdate handles PUT /users/:id
func (h *handler) handleUpdate(ctx *gin.Context) {
	id := ctx.Param("id")

	// bind partial update fields
	var fields *user.UpdateFields
	if err := ctx.ShouldBindJSON(&fields); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.userService.Update(id, fields)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, u)
}

// handleDelete handles DELETE /users/:id
func (h *handler) handleDelete(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := h.userService.Delete(id); err != nil {
		if errors.Is(err, user.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}

// El json que piden como respuesta del getSales
type SalesResponse struct {
	Metadata struct {
		Quantity    int     `json:"quantity"`
		Approved    int     `json:"approved"`
		Rejected    int     `json:"rejected"`
		Pending     int     `json:"pending"`
		TotalAmount float32 `json:"total_amount"`
	} `json:"metadata"`
	Results []*sales.Sales `json:"results"`
}

// handleCreate handles POST /sales
func (h *handler) handleCreateSales(ctx *gin.Context) {
	// request payload
	var req struct {
		UserID string  `json:"user_id"`
		Amount float32 `json:"amount"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s := &sales.Sales{
		UserID: req.UserID,
		Amount: req.Amount,
	}
	if err := h.salesService.Create(s); err != nil {
		if errors.Is(err, sales.ErrUserNotFound) || errors.Is(err, sales.ErrInvalidAmount) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("sale created", zap.Any("sale", s))
	ctx.JSON(http.StatusCreated, s)
}

// handleGetSales handles GET /sales
func (h *handler) handleGetSales(ctx *gin.Context) {
	user_id := ctx.Query("user_id")
	status := ctx.Query("status")

	if user_id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user_id es requerido"})
		return
	}

	salesList, err := h.salesService.GetSales(user_id, status)
	if err != nil {
		if errors.Is(err, sales.ErrInvalidStatus) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error("error obteniendo ventas",
			zap.String("user_id", user_id),
			zap.String("status", status),
			zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Calcular metadata
	var response SalesResponse

	// Asegurar que Results sea un array vacío si no hay ventas
	if salesList == nil {
		response.Results = []*sales.Sales{} // Inicializar como slice vacío
	} else {
		response.Results = salesList
	}

	response.Metadata.Quantity = len(response.Results) // Usar len del slice asignado

	for _, s := range response.Results { // Iterar sobre response.Results
		response.Metadata.TotalAmount += s.Amount
		switch s.Status {
		case "approved":
			response.Metadata.Approved++
		case "rejected":
			response.Metadata.Rejected++
		case "pending":
			response.Metadata.Pending++
		}
	}

	ctx.JSON(http.StatusOK, response)
}
// handleUpdateSales handles PATCH /sales
func (h *handler) handleUpdateSales(ctx *gin.Context) {
	sale_id := ctx.Param("id") // cambie aca para usar Param en lugar de Query

	if sale_id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "id es requerido"})
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validar que el status esté presente
	if req.Status == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "status es requerido"})
		return
	}

	// Actualizar la venta
	updatedSale, err := h.salesService.Update(sale_id, req.Status)
	if err != nil {
		if errors.Is(err, sales.ErrNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "sale not found"})
			return
		}
		if errors.Is(err, sales.ErrInvalidStatus) || errors.Is(err, sales.ErrEmptyID) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, sales.ErrInvalidTransition) {
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}

		h.logger.Error("error actualizando venta",
			zap.String("sale_id", sale_id),
			zap.String("status", req.Status),
			zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	h.logger.Info("sale updated", zap.Any("sale", updatedSale))
	ctx.JSON(http.StatusOK, updatedSale)
}
