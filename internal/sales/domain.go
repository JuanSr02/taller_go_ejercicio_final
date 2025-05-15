package sales

import "time"

// Sales represents a sale in the system with metadata for auditing and versioning.
type Sales struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Amount    float32   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   int       `json:"version"`
}

// UpdateFields represents the optional fields for updating a User.
// A nil pointer means “no change” for that field.
type UpdateFields struct {
	Status *string `json:"status"`
}
