package api

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/noelukwa/git-explorer/internal/explorer/api/handlers"
	"github.com/noelukwa/git-explorer/internal/explorer/service"
)

func SetupRoutes(intentService service.IntentService, repoService service.RemoteRepoService) *echo.Echo {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "SAMEORIGIN",
		HSTSMaxAge:            31536000,
		HSTSExcludeSubdomains: true,
	}))

	intentHandler := handlers.NewIntentHandler(intentService)

	e.POST("/intents", intentHandler.AddIntent)
	e.PUT("/intents/:id", intentHandler.UpdateIntent)
	e.GET("/intents/:id", intentHandler.FetchIntent)
	e.GET("/intents", intentHandler.FetchIntents)

	remoteRepoHandler := handlers.NewRemoteRepositoryHandler(repoService)
	e.GET("/repos/:name", remoteRepoHandler.FetchRepoInfo)
	return e
}
