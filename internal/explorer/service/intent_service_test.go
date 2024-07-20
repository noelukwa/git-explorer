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

// func TestCreateIntent(t *testing.T) {
// 	mockRepo := new(MockIntentRepository)
// 	service := service.NewIntentService(mockRepo)

// 	ctx := context.Background()
// 	repo := "test/repo"
// 	since := time.Now()

// 	mockRepo.On("SaveIntent", ctx, mock.AnythingOfType("*models.Intent")).Return(nil)

// 	intent, err := service.CreateIntent(ctx, repo, since)

// 	assert.NoError(t, err)
// 	assert.NotNil(t, intent)
// 	assert.Equal(t, repo, intent.Repository)
// 	assert.Equal(t, since, intent.Since)
// 	assert.True(t, intent.IsActive)
// 	mockRepo.AssertExpectations(t)
// }

// func TestGetIntent(t *testing.T) {
// 	mockRepo := new(MockIntentRepository)
// 	service := service.NewIntentService(mockRepo)

// 	ctx := context.Background()
// 	id := "test-id"
// 	expectedIntent := &models.Intent{ID: id}

// 	mockRepo.On("GetIntent", ctx, id).Return(expectedIntent, nil)

// 	intent, err := service.GetIntent(ctx, id)

// 	assert.NoError(t, err)
// 	assert.Equal(t, expectedIntent, intent)
// 	mockRepo.AssertExpectations(t)
// }

// func TestUpdateIntent(t *testing.T) {
// 	mockRepo := new(MockIntentRepository)
// 	service := service.NewIntentService(mockRepo)

// 	ctx := context.TODO()
// 	intentID := "123"
// 	since := time.Now()
// 	update := models.IntentUpdate{
// 		ID:       intentID,
// 		IsActive: true,
// 		Since:    &since,
// 	}

// 	originalIntent := &models.Intent{
// 		ID:       intentID,
// 		IsActive: false,
// 		Since:    time.Now().Add(-24 * time.Hour),
// 	}

// 	mockRepo.On("GetIntent", ctx, intentID).Return(originalIntent, nil)
// 	mockRepo.On("UpdateIntent", ctx, update).Return(nil)

// 	updatedIntent, err := service.UpdateIntent(ctx, update)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, updatedIntent)
// 	assert.Equal(t, update.IsActive, updatedIntent.IsActive)
// 	assert.Equal(t, *update.Since, updatedIntent.Since)

// 	mockRepo.AssertExpectations(t)
// }

// func TestCreateIntentError(t *testing.T) {
// 	mockRepo := new(MockIntentRepository)
// 	service := service.NewIntentService(mockRepo)

// 	ctx := context.Background()
// 	repo := "test/repo"
// 	since := time.Now()

// 	mockRepo.On("SaveIntent", ctx, mock.AnythingOfType("*models.Intent")).Return(errors.New("save error"))

// 	intent, err := service.CreateIntent(ctx, repo, since)

// 	assert.Error(t, err)
// 	assert.Nil(t, intent)
// 	mockRepo.AssertExpectations(t)
// }

// func TestGetIntentError(t *testing.T) {
// 	mockRepo := new(MockIntentRepository)
// 	service := service.NewIntentService(mockRepo)

// 	ctx := context.Background()
// 	id := "test-id"

// 	mockRepo.On("GetIntent", ctx, id).Return((*models.Intent)(nil), errors.New("not found"))

// 	intent, err := service.GetIntent(ctx, id)

// 	assert.Error(t, err)
// 	assert.Nil(t, intent)
// 	mockRepo.AssertExpectations(t)
// }

// func TestUpdateIntentGetError(t *testing.T) {
// 	mockRepo := new(MockIntentRepository)
// 	service := service.NewIntentService(mockRepo)

// 	ctx := context.Background()
// 	id := "test-id"
// 	update := models.IntentUpdate{IsActive: false, ID: id}

// 	mockRepo.On("GetIntent", ctx, id).Return((*models.Intent)(nil), errors.New("not found"))

// 	intent, err := service.UpdateIntent(ctx, update)

// 	assert.Error(t, err)
// 	assert.Nil(t, intent)
// 	mockRepo.AssertExpectations(t)
// }

// func TestUpdateIntentUpdateError(t *testing.T) {
// 	mockRepo := new(MockIntentRepository)
// 	service := service.NewIntentService(mockRepo)

// 	ctx := context.Background()
// 	id := "test-id"
// 	update := models.IntentUpdate{IsActive: false, ID: id}

// 	mockRepo.On("GetIntent", ctx, id).Return(&models.Intent{ID: id}, nil)
// 	mockRepo.On("UpdateIntent", ctx, update).Return(errors.New("update error"))

// 	intent, err := service.UpdateIntent(ctx, update)

// 	assert.Error(t, err)
// 	assert.Nil(t, intent)
// 	mockRepo.AssertExpectations(t)
// }
