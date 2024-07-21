package handlers

import (
	"net/http"

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

func (h *RemoteHandler) FetchToCommitters(c echo.Context) error {
	// limit := c.QueryParam("limit")
	// repo := c.QueryParam("repo")

	committers, err := h.remoteService.GetTopCommitters(c.Request().Context(), "", 1)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to top committers"})
	}
	return c.JSON(http.StatusOK, committers)
}

func (h *RemoteHandler) FetchRepoInfo(c echo.Context) error {
	repo := c.Param("repo")
	intent, err := h.remoteService.FindRepository(c.Request().Context(), repo)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update intent"})
	}

	return c.JSON(http.StatusOK, intent)
}
