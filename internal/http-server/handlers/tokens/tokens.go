package tokens

import (
	"context"
	"errors"
	resp "github.com/Gonnekone/ubiquitous-octo-fiesta/internal/lib/api/response"
	"github.com/Gonnekone/ubiquitous-octo-fiesta/internal/lib/jwt"
	"github.com/Gonnekone/ubiquitous-octo-fiesta/internal/lib/logger/sl"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

var (
	ErrEmptyGUID = errors.New("guid is empty")
)

//go:generate go run github.com/vektra/mockery/v2@v2.49.1 --name=RefreshTokenSaveDeleter
type RefreshTokenSaveDeleter interface {
	SaveRefreshToken(ctx context.Context, guid string, refreshToken string) error
	DeleteRefreshToken(ctx context.Context, guid string) error
}

func New(log *slog.Logger, refreshTokenSaveDeleter RefreshTokenSaveDeleter, jwtService *jwt.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.tokens.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		guid := r.URL.Query().Get("guid")
		if guid == "" {
			log.Error("invalid request", sl.Err(ErrEmptyGUID))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error(resp.InvalidRequest))

			return
		}

		ip := strings.Split(r.RemoteAddr, ":")[0]

		log.Debug("got request", slog.String("guid", guid), slog.String("ip", ip))

		tokens, err := jwtService.GenerateTokenPair(guid, ip, 3*time.Minute)
		if err != nil {
			log.Error("failed to generate tokens", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error(resp.InvalidRequest))

			return
		}

		log.Debug("tokens generated",
			slog.String("access_token", tokens.AccessToken),
			slog.String("refresh_token", tokens.RefreshToken),
		)

		hash, err := bcrypt.GenerateFromPassword([]byte(tokens.RefreshToken), bcrypt.DefaultCost)
		if err != nil {
			log.Error("failed to hash refresh token", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error(resp.InvalidRequest))

			return
		}

		if err := refreshTokenSaveDeleter.DeleteRefreshToken(r.Context(), guid); err != nil {
			log.Error("failed to delete refresh token hash", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error(resp.InvalidRequest))

			return
		}

		if err := refreshTokenSaveDeleter.SaveRefreshToken(r.Context(), guid, string(hash)); err != nil {
			log.Error("failed to save refresh token hash", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error(resp.InvalidRequest))

			return
		}

		render.JSON(w, r, tokens)
	}
}
