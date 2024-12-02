package refresh_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Gonnekone/ubiquitous-octo-fiesta/internal/http-server/handlers/refresh"
	"github.com/Gonnekone/ubiquitous-octo-fiesta/internal/http-server/handlers/refresh/mocks"
	"github.com/Gonnekone/ubiquitous-octo-fiesta/internal/lib/jwt"
	"github.com/Gonnekone/ubiquitous-octo-fiesta/internal/lib/logger/handlers/slogdiscard"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	secret  = "some_secret"
	some_ip = "127.0.0.1"
)

func TestRefreshHandler(t *testing.T) {
	refreshTokenProvider := mocks.NewRefreshTokenProvider(t)

	jwtService := jwt.New(secret)

	handler := refresh.New(slogdiscard.NewDiscardLogger(), refreshTokenProvider, jwtService)

	t.Run("Success", func(t *testing.T) {
		guid := uuid.New().String()
		tokens, err := jwtService.GenerateTokenPair(guid, some_ip, 3*time.Minute)
		require.NoError(t, err)

		hashedRefreshToken, err := bcrypt.GenerateFromPassword([]byte(tokens.RefreshToken), bcrypt.DefaultCost)
		require.NoError(t, err)

		refreshTokenProvider.On("SaveRefreshToken",
			context.Background(),
			guid,
			mock.AnythingOfType("string"),
		).
			Return(nil).
			Once()

		refreshTokenProvider.On("DeleteRefreshToken",
			context.Background(),
			guid,
		).
			Return(nil).
			Once()

		refreshTokenProvider.On("GetRefreshTokenHash",
			context.Background(),
			guid,
		).
			Return(string(hashedRefreshToken), nil).
			Once()

		input := fmt.Sprintf(`{"access_token": "%s", "refresh_token": "%s"}`, tokens.AccessToken, tokens.RefreshToken)

		req, err := http.NewRequest(http.MethodPost, "/refresh", bytes.NewReader([]byte(input)))
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		require.Equal(t, rr.Code, http.StatusOK)

		body := rr.Body.String()

		var resp jwt.Tokens

		require.NoError(t, json.Unmarshal([]byte(body), &resp))
	})

	t.Run("Empty access token", func(t *testing.T) {
		guid := uuid.New().String()
		tokens, err := jwtService.GenerateTokenPair(guid, some_ip, 3*time.Minute)
		require.NoError(t, err)

		input := fmt.Sprintf(`{"access_token": "%s", "refresh_token": "%s"}`, "", tokens.RefreshToken)

		req, err := http.NewRequest(http.MethodPost, "/refresh", bytes.NewReader([]byte(input)))
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		require.Equal(t, rr.Code, http.StatusBadRequest)
	})

	t.Run("Empty refresh token", func(t *testing.T) {
		guid := uuid.New().String()
		tokens, err := jwtService.GenerateTokenPair(guid, some_ip, 3*time.Minute)
		require.NoError(t, err)

		input := fmt.Sprintf(`{"access_token": "%s", "refresh_token": "%s"}`, tokens.AccessToken, "")

		req, err := http.NewRequest(http.MethodPost, "/refresh", bytes.NewReader([]byte(input)))
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		require.Equal(t, rr.Code, http.StatusBadRequest)
	})

	t.Run("Wrong access token", func(t *testing.T) {
		guid := uuid.New().String()
		tokens, err := jwtService.GenerateTokenPair(guid, some_ip, 3*time.Minute)
		require.NoError(t, err)

		input := fmt.Sprintf(`{"access_token": "%s", "refresh_token": "%s"}`, tokens.AccessToken[:len(tokens.AccessToken)-1], tokens.RefreshToken)

		req, err := http.NewRequest(http.MethodPost, "/refresh", bytes.NewReader([]byte(input)))
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		require.Equal(t, rr.Code, http.StatusBadRequest)
	})

	t.Run("Not actual refresh token", func(t *testing.T) {
		guid := uuid.New().String()
		tokens, err := jwtService.GenerateTokenPair(guid, some_ip, 3*time.Minute)
		require.NoError(t, err)

		refreshTokenProvider.On("GetRefreshTokenHash",
			context.Background(),
			guid,
		).
			Return("", nil).
			Once()

		input := fmt.Sprintf(`{"access_token": "%s", "refresh_token": "%s"}`, tokens.AccessToken, tokens.RefreshToken)

		req, err := http.NewRequest(http.MethodPost, "/refresh", bytes.NewReader([]byte(input)))
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		require.Equal(t, rr.Code, http.StatusBadRequest)
	})
}
