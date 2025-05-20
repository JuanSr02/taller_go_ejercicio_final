package sales

import "errors"

// ErrNotFound is returned when a sale with the given ID is not found.
var ErrNotFound = errors.New("sale not found")

// ErrEmptyID is returned when trying to store a sale with an empty ID.
var ErrEmptyID = errors.New("empty sale ID")

// ErrInvalidAmount is returned when trying to store a sale with an invalid Amount.
var ErrInvalidAmount = errors.New("invalid amount")

// ErrUserNotFound is returned when a user with the given ID is not found.
var ErrUserNotFound = errors.New("user not found")

// Storage is the main interface for our storage layer.
type Storage interface {
	Set(sales *Sales) error
	Read(id string) (*Sales, error)
	Delete(id string) error
	GetAll(user_id string) ([]*Sales, error)
	GetByStatus(user_id, status string) ([]*Sales, error)
}

// LocalStorage provides an in-memory implementation for storing sales.
type LocalStorage struct {
	m map[string]*Sales
}

// NewLocalStorage instantiates a new LocalStorage with an empty map.
func NewLocalStorage() *LocalStorage {
	return &LocalStorage{
		m: map[string]*Sales{},
	}
}

// Set stores or updates a sale in the local storage.
// Returns ErrEmptyID if the sale has an empty ID.
func (l *LocalStorage) Set(sales *Sales) error {
	if sales.ID == "" {
		return ErrEmptyID
	}

	l.m[sales.ID] = sales
	return nil
}

// Read retrieves a sale from the local storage by ID.
// Returns ErrNotFound if the sale is not found.
func (l *LocalStorage) Read(id string) (*Sales, error) {
	s, ok := l.m[id]
	if !ok {
		return nil, ErrNotFound
	}

	return s, nil
}

// Delete removes a sale from the local storage by ID.
// Returns ErrNotFound if the sale does not exist.
func (l *LocalStorage) Delete(id string) error {
	_, err := l.Read(id)
	if err != nil {
		return err
	}

	delete(l.m, id)
	return nil
}

// GetAll retorna todas las ventas de un usuario dado su ID
func (l *LocalStorage) GetAll(user_id string) ([]*Sales, error) {
	var sales []*Sales
	for _, s := range l.m {
		if s.UserID == user_id {
			sales = append(sales, s)
		}
	}
	return sales, nil
}

// GetByStatus returns todas las ventas de un usuario dado su ID y filtrando por estado
func (l *LocalStorage) GetByStatus(user_id, status string) ([]*Sales, error) {
	var sales []*Sales
	for _, s := range l.m {
		if s.UserID == user_id && s.Status == status {
			sales = append(sales, s)
		}
	}
	return sales, nil
}
