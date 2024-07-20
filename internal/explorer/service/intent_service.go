package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/noelukwa/git-explorer/internal/events"
	"github.com/noelukwa/git-explorer/internal/explorer/models"
	"github.com/noelukwa/git-explorer/internal/explorer/repository"
	"github.com/noelukwa/git-explorer/internal/pkg/messaging"
)

var (
	ErrInvalidRepository = errors.New("invalid repo, only accept <owner>/<repo> format")
	ErrExistingIntent    = errors.New("repo intent already booked")
)

type IntentService interface {
	CreateIntent(ctx context.Context, repo string, since time.Time) (*models.Intent, error)
	GetIntentById(ctx context.Context, id uuid.UUID) (*models.Intent, error)
	UpdateIntent(ctx context.Context, update models.IntentUpdate) (*models.Intent, error)
	GetIntents(ctx context.Context, isActive bool) ([]*models.Intent, error)
}

type intentService struct {
	repo          repository.IntentRepository
	messageClient messaging.Client
}

func NewIntentService(repo repository.IntentRepository, mClient messaging.Client) IntentService {
	return &intentService{repo: repo, messageClient: mClient}
}

func (i *intentService) CreateIntent(ctx context.Context, repo string, since time.Time) (*models.Intent, error) {
	if len(strings.Split(repo, "/")) < 2 {
		return nil, ErrInvalidRepository
	}

	if since.IsZero() {
		since = time.Now()
	}

	oldIntent, err := i.repo.GetIntentByRepo(ctx, repo)
	if err != nil {
		return nil, err
	}

	if oldIntent != nil {
		return nil, ErrExistingIntent
	}
	uid, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	intent := &models.Intent{
		ID:         uid,
		Repository: repo,
		Since:      since,
		CreatedAt:  time.Now(),
		IsActive:   true,
	}

	err = i.repo.SaveIntent(ctx, intent)
	if err != nil {
		return nil, err
	}

	return intent, nil
}

func (i *intentService) GetIntentById(ctx context.Context, id uuid.UUID) (*models.Intent, error) {
	return i.repo.GetIntentById(ctx, id)
}

func (i *intentService) GetIntents(ctx context.Context, isActive bool) ([]*models.Intent, error) {
	return i.repo.GetIntents(ctx, repository.IntentFilter{
		IsActive: isActive,
	})
}

func (i *intentService) UpdateIntent(ctx context.Context, update models.IntentUpdate) (*models.Intent, error) {
	intent, err := i.repo.GetIntentById(ctx, update.ID)
	if err != nil {
		return nil, err
	}

	err = i.repo.UpdateIntent(ctx, &update)
	if err != nil {
		return nil, err
	}

	if !update.Since.IsZero() {
		intent.Since = *update.Since
	}

	return intent, nil
}

func (i *intentService) sendNewIntentEvent(intent *models.Intent) error {
	eventJSON, err := json.Marshal(intent)
	if err != nil {
		return fmt.Errorf("error marshaling intent event: %w", err)
	}

	event := messaging.Event{
		Subject: events.NEW_INTENT,
		Data:    eventJSON,
	}

	err = i.messageClient.Publish(event)
	if err != nil {
		return fmt.Errorf("error publishing intent event: %w", err)
	}

	log.Println("published new intent yo!!")
	return nil
}
