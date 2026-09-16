package configuration

import (
	"context"
	"fmt"

	"github.com/anderson-reinaldo/go-explicAI/internal/infrastructure/api"
	"github.com/anderson-reinaldo/go-explicAI/internal/infrastructure/log"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Application struct {
	server *echo.Echo
	config *viper.Viper
}

func NewApplication(config *viper.Viper) *Application {
	server := echo.New()
	server.HideBanner = true
	server.HidePort = true

	logger := log.StartLog()
	initMiddlewares(server, logger)
	return &Application{
		server: server,
		config: config,
	}
}

func (a *Application) Start() {
	a.registerControllers()
	ctx := context.Background()
	log.LogInfo(ctx, fmt.Sprintf("%s is starting on %s", a.config.GetString("app.name"), a.config.GetString("server.host")))
	log.LogError(ctx, "server fatal error", a.server.Start(a.config.GetString("server.host")))
}

func (a *Application) registerControllers() {
	api.NewExplicaServer().Register(a.server)
}

func initMiddlewares(server *echo.Echo, logger *zap.Logger) {
	server.Use(middleware.Recover())
	server.Use(loggerMiddleware(logger))
}

func loggerMiddleware(logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			requestID := uuid.NewString()
			ctxLogger := logger.With(
				zap.String("requestId", requestID),
				zap.String("remoteIp", c.RealIP()),
				zap.String("path", c.Path()),
			)

			ctx := context.WithValue(c.Request().Context(), "logger", ctxLogger)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
