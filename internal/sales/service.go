package sales

import (
	"errors"
	"math/rand"
	"time"

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
func (s *Service) Create(sales *Sales, user_id string, amount float32) error {
	// Checks if the ID given is from a User that exits, else it will give an error
	// falta get de user id
	sales.ID = uuid.NewString()
	sales.UserID = user_id
	if amount != 0 {
		sales.Amount = amount
	} else {
		s.logger.Error("failed to set sale", zap.Error(ErrInvalidAmount), zap.Any("sales", sales))
		return ErrInvalidAmount
	}
	sales.Status = status_options[rand.Intn(len(status_options))]

	now := time.Now()
	sales.CreatedAt = now
	sales.UpdatedAt = now
	sales.Version = 1

	if err := s.storage.Set(sales); err != nil {
		s.logger.Error("failed to set sale", zap.Error(err), zap.Any("sales", sales))
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
