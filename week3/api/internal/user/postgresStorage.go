package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(ctx context.Context, databaseURL string) (*PostgresStorage, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to connect to platform database: %w", err)
	}

	return &PostgresStorage{
		pool: pool,
	}, nil
}

func (s *PostgresStorage) Close() {
	s.pool.Close()
}

func (s *PostgresStorage) Create(ctx context.Context, u User) error {
	const query = `
		INSERT INTO users (
			id,
			username,
			password_hash,
			role,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := s.pool.Exec(
		ctx,
		query,
		u.ID,
		u.Username,
		u.PasswordHash,
		u.Role,
		u.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrUsernameExists
		}

		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (s *PostgresStorage) GetByUsername(
	ctx context.Context,
	username string,
) (*User, error) {
	const query = `
		SELECT
			id,
			username,
			password_hash,
			role,
			created_at
		FROM users
		WHERE username = $1
	`

	var u User

	err := s.pool.QueryRow(ctx, query, username).Scan(
		&u.ID,
		&u.Username,
		&u.PasswordHash,
		&u.Role,
		&u.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &u, nil
}

func (s *PostgresStorage) GetByID(
	ctx context.Context,
	id string,
) (*User, error) {
	const query = `
		SELECT
			id,
			username,
			password_hash,
			role,
			created_at
		FROM users
		WHERE id = $1
	`

	var u User

	err := s.pool.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.Username,
		&u.PasswordHash,
		&u.Role,
		&u.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &u, nil
}