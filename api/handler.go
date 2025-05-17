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
	salesService *sales.Service // Add this
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

func (h *handler) handleCreateSale(ctx *gin.Context) {
	// TODO
}

func (h *handler) handleUpdateSale(ctx *gin.Context) {
	// TODO
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
	response.Results = salesList
	response.Metadata.Quantity = len(salesList)

	for _, s := range salesList {
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
