package users

import (
	"context"

	"github.com/google/uuid"
)

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) CreateUser(ctx context.Context, username string) (*User, error) {

	user, err := s.repo.GetUserByUsername(ctx, username)

	if err != nil {
		return nil, err
	}

	if user != nil {
		return nil, ErrUsernameTaken
	}

	user, err = s.repo.CreateUser(ctx, username)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	user, err := s.repo.GetUserByUsername(ctx, username)

	if err != nil {
		return nil, err
	}

	return user, err
}

func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	user, err := s.repo.GetUserByID(ctx, id)

	if err != nil {
		return nil, err
	}

	return user, nil
}
