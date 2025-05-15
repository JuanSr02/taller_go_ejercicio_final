package sales

import (
	"github.com/google/uuid"
	"go.uber.org/zap"
	"math/rand"
	"time"
)

// Service provides high-level sales management operations on a LocalStorage backend.
type Service struct {
	// storage is the underlying persistence for Sales entities.
	storage Storage

	// logger is our observability component to log.
	logger *zap.Logger
}

// 0: pending, 1: approver, 2: rejected
// La forma mas facil que me salio pa que elija aleatoriamente en el create jeje
var status_options = []string{"pending", "approved", "rejected"}

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

// Get retrieves a user by its ID.
// Returns ErrNotFound if no user exists with the given ID.
func (s *Service) Get(id string) (*User, error) {
	return s.storage.Read(id)
}

// Update modifies an existing user's data.
// It updates Name, Address, NickName, sets UpdatedAt to now and increments Version.
// Returns ErrNotFound if the user does not exist, or ErrEmptyID if user.ID is empty.
func (s *Service) Update(id string, user *UpdateFields) (*User, error) {
	existing, err := s.storage.Read(id)
	if err != nil {
		return nil, err
	}

	if user.Name != nil {
		existing.Name = *user.Name
	}

	if user.Address != nil {
		existing.Address = *user.Address
	}

	if user.NickName != nil {
		existing.NickName = *user.NickName
	}

	existing.UpdatedAt = time.Now()
	existing.Version++

	if err := s.storage.Set(existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// Delete removes a user from the system by its ID.
// Returns ErrNotFound if the user does not exist.
func (s *Service) Delete(id string) error {
	return s.storage.Delete(id)
}
