package tests

import (
	"ej_final/internal/sales"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestService_Create_Simple(t *testing.T) {
	mockHandler := http.NewServeMux()
	mockServer := httptest.NewServer(mockHandler)
	defer mockServer.Close()

	s := sales.NewService(sales.NewLocalStorage(), zap.NewNop(), mockServer.URL)

	input := &sales.Sales{
		UserID: "Pepe",
		Amount: 1.0,
	}

	err := s.Create(input)

	require.EqualError(t, err, sales.ErrUserNotFound.Error())
	require.NotEmpty(t, input.UserID)
	require.NotEmpty(t, input.Amount)
}
