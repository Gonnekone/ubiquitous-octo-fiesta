package refresh

import (
	"context"
	"errors"
	resp "github.com/Gonnekone/ubiquitous-octo-fiesta/internal/lib/api/response"
	"github.com/Gonnekone/ubiquitous-octo-fiesta/internal/lib/email"
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

const emailToSend = "opacha2018@yandex.ru"

var (
	ErrEmptyAccessToken  = errors.New("access token is empty")
	ErrEmptyRefreshToken = errors.New("access token is empty")
)

//go:generate go run github.com/vektra/mockery/v2@v2.49.1 --name=RefreshTokenProvider
type RefreshTokenProvider interface {
	SaveRefreshToken(ctx context.Context, guid string, refreshToken string) error
	DeleteRefreshToken(ctx context.Context, guid string) error
	GetRefreshTokenHash(ctx context.Context, guid string) (string, error)
}

type Request struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func New(log *slog.Logger, refreshTokenProvider RefreshTokenProvider, jwtService *jwt.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.refresh.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		if req.AccessToken == "" {
			log.Error("invalid request", sl.Err(ErrEmptyAccessToken))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error(resp.InvalidRequest))

			return
		}

		if req.RefreshToken == "" {
			log.Error("invalid request", sl.Err(ErrEmptyRefreshToken))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error(resp.InvalidRequest))

			return
		}

		ip := strings.Split(r.RemoteAddr, ":")[0]

		log.Debug("got request",
			slog.String("access_token", req.AccessToken),
			slog.String("refresh_token", req.RefreshToken),
			slog.String("ip", ip),
		)

		accessIp, accessGuid, err := jwtService.ValidateAccessToken(req.AccessToken)
		if err != nil {
			log.Error("failed to validate access token", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error(resp.InvalidRequest))

			return
		}

		log.Debug("got from access token", slog.String("ip", accessIp), slog.String("guid", accessGuid))

		if accessIp != ip {
			log.Debug("ip changed, sending email", slog.String("access_ip", accessIp), slog.String("ip", ip))

			var err error

			go func() {
				err = email.SendEmailWarning(emailToSend)
			}()

			if err != nil {
				log.Error("failed to send email", sl.Err(err))
			}
		}

		refreshTokenHash, err := refreshTokenProvider.GetRefreshTokenHash(r.Context(), accessGuid)
		if err != nil {
			log.Error("failed to get refresh token hash", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error(resp.InvalidRequest))

			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(refreshTokenHash), []byte(req.RefreshToken)); err != nil {
			log.Error("refresh token does not exist", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error(resp.InvalidRequest))

			return
		}

		err = jwtService.ValidateRefreshToken(req.RefreshToken, req.AccessToken, accessIp)
		if err != nil {
			log.Error("failed to validate refresh token", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error(resp.InvalidRequest))

			return
		}

		if err := refreshTokenProvider.DeleteRefreshToken(r.Context(), accessGuid); err != nil {
			log.Error("failed to delete refresh token hash", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error(resp.InvalidRequest))

			return
		}

		tokens, err := jwtService.GenerateTokenPair(accessGuid, ip, 3*time.Minute)
		if err != nil {
			log.Error("failed to generate tokens", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error(resp.InvalidRequest))

			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(tokens.RefreshToken), bcrypt.DefaultCost)
		if err != nil {
			log.Error("failed to hash refresh token", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error(resp.InvalidRequest))

			return
		}

		if err := refreshTokenProvider.SaveRefreshToken(r.Context(), accessGuid, string(hash)); err != nil {
			log.Error("failed to save refresh token hash", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error(resp.InvalidRequest))

			return
		}

		render.JSON(w, r, tokens)
	}
}
