package users

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

const defaultListTimeout = time.Second * 3

type repository interface {
	ListUsers(ctx context.Context) []User
	FetchUser(ctx context.Context, id string) (*User, error)
}

type Service struct {
	logger     *slog.Logger
	repository repository
}

func NewService(logger *slog.Logger, repo repository) *Service {
	return &Service{
		logger:     logger.WithGroup("users-service"),
		repository: repo,
	}
}

func (s *Service) ListUsers(ctx context.Context) []User {
	ctx, cancel := context.WithTimeout(ctx, defaultListTimeout)
	defer cancel()
	return s.repository.ListUsers(ctx)
}

func (s *Service) FetchUser(ctx context.Context, id string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultListTimeout)
	defer cancel()

	u, err := s.repository.FetchUser(ctx, id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("could not fetch user id '%s' from repository: %w", id, err)
	}
	return u, nil
}
