package api

import (
	"ej_final/internal/sales"
	"ej_final/internal/user"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// InitRoutes registers all user CRUD endpoints on the given Gin engine.
// It initializes the storage, service, and handler, then binds each HTTP
// method and path to the appropriate handler function.
func InitRoutes(e *gin.Engine) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Inicializar user service
	userStorage := user.NewLocalStorage()
	userService := user.NewService(userStorage, logger)

	// Inicializar sales service
	salesStorage := sales.NewLocalStorage()
	salesService := sales.NewService(salesStorage, logger)

	h := handler{
		userService:  userService,
		salesService: salesService,
		logger:       logger,
	}

	// Existing routes
	e.POST("/users", h.handleCreate)
	e.GET("/users/:id", h.handleRead)
	e.PATCH("/users/:id", h.handleUpdate)
	e.DELETE("/users/:id", h.handleDelete)

	// Add new route
	e.POST("/sales", h.handleCreateSales) // hacelo @luuLoyola
	e.GET("/sales", h.handleGetSales)
	e.PATCH("/sales/:id", h.handleUpdateSales) // hacelo @fabriBauer

	e.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
}
