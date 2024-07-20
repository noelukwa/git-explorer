package inmem

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/noelukwa/git-explorer/internal/explorer/models"
	"github.com/noelukwa/git-explorer/internal/explorer/repository"
)

type IntentRepository struct {
	intents map[string]*models.Intent
	mu      sync.RWMutex
}

func (r *IntentRepository) GetIntentById(ctx context.Context, id uuid.UUID) (*models.Intent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, intent := range r.intents {
		if intent.ID == id {
			return intent, nil
		}
	}

	return nil, nil
}

func (r *IntentRepository) GetIntentByRepo(ctx context.Context, repo string) (*models.Intent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, intent := range r.intents {
		if intent.Repository == repo {
			return intent, nil
		}
	}

	return nil, nil
}

func (r *IntentRepository) GetIntents(ctx context.Context, filter repository.IntentFilter) ([]*models.Intent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*models.Intent

	for _, intent := range r.intents {
		if filter.IsActive == intent.IsActive {
			result = append(result, intent)
		}
	}

	return result, nil
}

func newIntentRepository() *IntentRepository {
	return &IntentRepository{
		intents: make(map[string]*models.Intent),
	}
}

func (r *IntentRepository) SaveIntent(ctx context.Context, intent *models.Intent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if intent.ID == uuid.Nil {
		return errors.New("intent ID cannot be empty")
	}

	r.intents[intent.ID.String()] = intent
	return nil
}

func (r *IntentRepository) UpdateIntent(ctx context.Context, update *models.IntentUpdate) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	intent, exists := r.intents[update.ID.String()]
	if !exists {
		return errors.New("intent not found")
	}

	intent.IsActive = update.IsActive
	if update.Since != nil {
		intent.Since = *update.Since
	}

	return nil
}
