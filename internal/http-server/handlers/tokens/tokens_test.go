package tokens_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Gonnekone/ubiquitous-octo-fiesta/internal/http-server/handlers/tokens"
	"github.com/Gonnekone/ubiquitous-octo-fiesta/internal/http-server/handlers/tokens/mocks"
	"github.com/Gonnekone/ubiquitous-octo-fiesta/internal/lib/jwt"
	"github.com/Gonnekone/ubiquitous-octo-fiesta/internal/lib/logger/handlers/slogdiscard"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

const secret = "some_secret"

func TestTokensHandler(t *testing.T) {
	refreshTokenSaveDeleter := mocks.NewRefreshTokenSaveDeleter(t)

	jwtService := jwt.New(secret)

	handler := tokens.New(slogdiscard.NewDiscardLogger(), refreshTokenSaveDeleter, jwtService)

	t.Run("Success", func(t *testing.T) {
		guid := uuid.New().String()
		refreshTokenSaveDeleter.On("DeleteRefreshToken",
			context.Background(),
			guid,
		).
			Return(nil).
			Once()

		refreshTokenSaveDeleter.On("SaveRefreshToken",
			context.Background(),
			guid,
			mock.AnythingOfType("string"),
		).
			Return(nil).
			Once()

		url := fmt.Sprintf("/get-tokens?guid=%s", guid)

		req, err := http.NewRequest(http.MethodGet, url, bytes.NewReader([]byte{}))
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		require.Equal(t, rr.Code, http.StatusOK)

		body := rr.Body.String()
		var resp jwt.Tokens

		require.NoError(t, json.Unmarshal([]byte(body), &resp))
	})

	t.Run("Empty guid", func(t *testing.T) {
		url := fmt.Sprintf("/get-tokens?guid=%s", "")

		req, err := http.NewRequest(http.MethodGet, url, bytes.NewReader([]byte{}))
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		require.Equal(t, rr.Code, http.StatusBadRequest)
	})
}
