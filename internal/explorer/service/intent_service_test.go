package service_test

import (
	"context"

	"github.com/google/uuid"
	"github.com/noelukwa/git-explorer/internal/explorer/models"
	"github.com/noelukwa/git-explorer/internal/explorer/repository"

	"github.com/stretchr/testify/mock"
)

type MockIntentRepository struct {
	mock.Mock
}

func (m *MockIntentRepository) SaveIntent(ctx context.Context, intent *models.Intent) error {
	args := m.Called(ctx, intent)
	return args.Error(0)
}

func (m *MockIntentRepository) GetIntent(ctx context.Context, id string) (*models.Intent, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Intent), args.Error(1)
}

func (m *MockIntentRepository) FindIntents(ctx context.Context, pagination repository.Pagination) (repository.PaginatedResponse[models.Intent], error) {
	args := m.Called(ctx, pagination)
	return args.Get(0).(repository.PaginatedResponse[models.Intent]), args.Error(1)
}

func (m *MockIntentRepository) UpdateIntent(ctx context.Context, intent *models.IntentUpdate) error {
	args := m.Called(ctx, intent)
	return args.Error(0)
}

func (m *MockIntentRepository) DeleteIntent(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
