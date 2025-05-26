package sales

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service provides high-level sales management operations on a LocalStorage backend.
type Service struct {
	// storage is the underlying persistence for Sales entities.
	storage Storage

	// logger is our observability component to log.
	logger *zap.Logger
}

// 0: pending, 1: approved, 2: rejected
// La forma mas facil que me salio pa que elija aleatoriamente en el create jeje
var status_options = []string{"pending", "approved", "rejected"}

// Para tener el error personalizado jeee
var ErrInvalidStatus = errors.New("invalid status")
var ErrInvalidTransition = errors.New("invalid status transition")

// NewService creates a new Service.
func NewService(storage Storage, logger *zap.Logger) *Service {
	if logger == nil {
		logger, _ = zap.NewProduction()
		defer logger.Sync() // flushes buffer, if any
	}

	return &Service{
		storage: storage,
		logger:  logger,
	}
}

// Create adds a brand-new sale to the system.
// It sets CreatedAt and UpdatedAt to the current time and initializes Version to 1.
// Returns ErrEmptyID if sales.ID is empty.
func (s *Service) Create(sales *Sales) error {
	// Checks if the ID given is from a User that exits, else it will give an error
	client := resty.New()

	resp, err := client.R().
		Get(fmt.Sprintf("http://localhost:8080/users/%s", sales.UserID))

	if err != nil {
		s.logger.Error("Ocurrio un error al buscar el ID del usuario", zap.Error(err))
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		s.logger.Error("ID de Usuario dado no existe", zap.Error(err))
		return ErrUserNotFound
	}

	sales.ID = uuid.NewString()
	if sales.Amount <= 0 {
		s.logger.Error("Amount no puede ser un valor menor o igual a 0", zap.Error(ErrInvalidAmount), zap.Any("sales", sales))
		return ErrInvalidAmount
	}
	sales.Status = status_options[rand.Intn(len(status_options))]

	now := time.Now()
	sales.CreatedAt = now
	sales.UpdatedAt = now
	sales.Version = 1

	if err := s.storage.Set(sales); err != nil {
		s.logger.Error("Error al crear la venta", zap.Error(err), zap.Any("sales", sales))
		return err
	}
	return nil
}

func (s *Service) GetSales(user_id, status string) ([]*Sales, error) {
	// Validar estado si fue dado
	if status != "" {
		validStatus := false
		for _, st := range status_options {
			if status == st {
				validStatus = true
				break
			}
		}
		if !validStatus {
			s.logger.Error("El estado dado es invalido", zap.String("status", status))
			return nil, ErrInvalidStatus
		}
		sales, err := s.storage.GetByStatus(user_id, status)
		if err != nil {
			s.logger.Error("Error obteniendo ventas por estado",
				zap.String("user_id", user_id),
				zap.String("status", status),
				zap.Error(err))
			return nil, err
		}
		return sales, nil
	}

	// Si no se dio un estado, obtener todas las ventas
	sales, err := s.storage.GetAll(user_id)
	if err != nil {
		s.logger.Error("Error obteniendo todas las ventas",
			zap.String("user_id", user_id),
			zap.Error(err))
		return nil, err
	}
	return sales, nil
}

func (s *Service) Update(saleID string, newStatus string) (*Sales, error) {
	// Validar que el ID no esté vacío
	if saleID == "" {
		s.logger.Error("ID de venta está vacío")
		return nil, ErrEmptyID
	}

	// Validar que el nuevo estado sea válido
	validStatus := false
	for _, st := range status_options {
		if newStatus == st {
			validStatus = true
			break
		}
	}
	if !validStatus {
		s.logger.Error("El estado dado es inválido", zap.String("status", newStatus))
		return nil, ErrInvalidStatus
	}

	// Obtener la venta actual
	sale, err := s.storage.Read(saleID)
	if err != nil {
		s.logger.Error("Error obteniendo venta para actualizar",
			zap.String("sale_id", saleID),
			zap.Error(err))
		return nil, err
	}

	// Validar transición: solo se puede cambiar desde "pending"
	if sale.Status != "pending" {
		s.logger.Error("Transición inválida: la venta no está en estado pending",
			zap.String("sale_id", saleID),
			zap.String("current_status", sale.Status),
			zap.String("new_status", newStatus))
		return nil, ErrInvalidTransition
	}

	// Validar que solo se pueda cambiar a "approved" o "rejected"
	if newStatus != "approved" && newStatus != "rejected" {
		s.logger.Error("Solo se puede cambiar de pending a approved o rejected",
			zap.String("sale_id", saleID),
			zap.String("new_status", newStatus))
		return nil, ErrInvalidTransition
	}

	// Actualizar la venta
	sale.Status = newStatus
	sale.UpdatedAt = time.Now()
	sale.Version++

	// Guardar la venta actualizada
	if err := s.storage.Set(sale); err != nil {
		s.logger.Error("Error actualizando la venta",
			zap.String("sale_id", saleID),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("Venta actualizada exitosamente",
		zap.String("sale_id", saleID),
		zap.String("new_status", newStatus))

	return sale, nil
}
