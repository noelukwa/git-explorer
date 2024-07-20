package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/noelukwa/git-explorer/internal/explorer/models"
	"github.com/noelukwa/git-explorer/internal/explorer/service"
)

type Since time.Time

func (ct *Since) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	*ct = Since(t)
	return nil
}

type IntentHandler struct {
	intentService service.IntentService
	validator     *validator.Validate
}

func NewIntentHandler(intentService service.IntentService) *IntentHandler {
	return &IntentHandler{
		intentService: intentService,
		validator:     validator.New(),
	}
}

type AddIntentRequest struct {
	Repo  string `json:"repo" validate:"required"`
	Since Since  `json:"since" validate:"required"`
}

func (h *IntentHandler) AddIntent(c echo.Context) error {
	var request AddIntentRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := h.validator.Struct(request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	intent, err := h.intentService.CreateIntent(
		c.Request().Context(),
		request.Repo,
		time.Time(request.Since),
	)
	if err != nil {
		if errors.Is(err, service.ErrInvalidRepository) || errors.Is(err, service.ErrExistingIntent) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		log.Printf("error: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to add intent"})
	}

	return c.JSON(http.StatusCreated, intent)
}

type UpdateIntentRequest struct {
	IsActive bool  `json:"is_active"`
	Since    Since `json:"since"`
}

func (h *IntentHandler) UpdateIntent(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid intent ID"})
	}

	var request UpdateIntentRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body!!"})
	}

	if err := h.validator.Struct(request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	since := time.Time(request.Since)
	intentUpdate := models.IntentUpdate{
		ID:       id,
		IsActive: request.IsActive,
		Since:    &since,
	}

	intent, err := h.intentService.UpdateIntent(c.Request().Context(), intentUpdate)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update intent"})
	}

	return c.JSON(http.StatusOK, intent)
}

func (h *IntentHandler) FetchIntent(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid intent ID"})
	}

	intent, err := h.intentService.GetIntentById(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update intent"})
	}

	return c.JSON(http.StatusOK, intent)
}

func (h *IntentHandler) FetchIntents(c echo.Context) error {

	var isActive bool

	flag := c.QueryParam("is_active")
	if flag != "" {
		boolValue, err := strconv.ParseBool(flag)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid is_active parameter"})
		}
		isActive = boolValue
	}

	intents, err := h.intentService.GetIntents(c.Request().Context(), isActive)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch intents"})
	}
	return c.JSON(http.StatusOK, intents)
}
