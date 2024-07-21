package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/noelukwa/git-explorer/internal/explorer/service"
)

type RemoteHandler struct {
	remoteService service.RemoteRepoService
	validator     *validator.Validate
}

func NewRemoteRepositoryHandler(remoteService service.RemoteRepoService) *RemoteHandler {
	return &RemoteHandler{
		remoteService: remoteService,
		validator:     validator.New(),
	}
}

func (h *RemoteHandler) FetchTopCommitters(c echo.Context) error {
	limitStr := c.QueryParam("limit")
	repo := c.QueryParam("repo")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		// If the limit is not a valid integer, return a bad request error
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid limit parameter"})
	}

	committers, err := h.remoteService.GetTopCommitters(c.Request().Context(), repo, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get top committers"})
	}
	return c.JSON(http.StatusOK, committers)
}

func (h *RemoteHandler) FetchRepoInfo(c echo.Context) error {
	repo := c.Param("name")
	intent, err := h.remoteService.FindRepository(c.Request().Context(), repo)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update intent"})
	}

	return c.JSON(http.StatusOK, intent)
}
