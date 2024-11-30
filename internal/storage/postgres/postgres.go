package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log/slog"
)

type Storage struct {
	logger *slog.Logger
	db     *pgx.Conn
}

func New(logger *slog.Logger, dsn string) (*Storage, error) {
	const op = "storage.postgres.New"

	log := logger.With(
		slog.String("op", op),
	)

	log.Debug("connecting to postgres", slog.String("dsn", dsn))

	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Debug("creating table refresh_tokens")

	_, err = conn.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS refresh_tokens (
		id TEXT PRIMARY KEY,
		refresh_token TEXT UNIQUE NOT NULL
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: conn, logger: logger}, nil
}

func (s *Storage) SaveRefreshToken(ctx context.Context, guid string, refreshToken string) error {
	_, err := s.db.Exec(ctx,
		"INSERT INTO refresh_tokens(id, refresh_token) VALUES($1, $2)",
		guid, refreshToken)

	if err != nil {
		return fmt.Errorf("insert refresh token: %w", err)
	}

	return nil
}

func (s *Storage) DeleteRefreshToken(ctx context.Context, guid string) error {
	_, err := s.db.Exec(ctx, "DELETE FROM refresh_tokens WHERE id = $1", guid)

	if err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}

	return err
}

func (s *Storage) TokenExists(ctx context.Context, guid string) (string, error) {
	rows, err := s.db.Query(ctx, "SELECT refresh_token FROM refresh_tokens WHERE id = $1", guid)
	defer rows.Close()
	if err != nil {
		return "", fmt.Errorf("select refresh token: %w", err)
	}

	var hash string

	if rows.Next() {
		err = rows.Scan(&hash)
		if err != nil {
			return "", fmt.Errorf("scan refresh token: %w", err)
		}
	}

	return hash, nil
}

func (s *Storage) Close() {
	s.logger.Debug("closing postgres connection")
	s.db.Close(context.Background())
}
