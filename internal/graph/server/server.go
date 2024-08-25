package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/lingticio/gateway/internal/configs"
	"github.com/lingticio/gateway/internal/graph/openai"
	"github.com/lingticio/gateway/internal/graph/server/middlewares"
	"github.com/nekomeowww/xo/logger"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Modules() fx.Option {
	return fx.Options(
		fx.Provide(openai.NewGraphQLHandler()),
		fx.Provide(NewServer()),
	)
}

type NewServerParams struct {
	fx.In

	Lifecycle fx.Lifecycle
	Logger    *logger.Logger
	Config    *configs.Config

	OpenAIHandler *openai.GraphQLHandler
}

type Server struct {
	*http.Server
}

func NewServer() func(params NewServerParams) *Server {
	return func(params NewServerParams) *Server {
		e := echo.New()

		e.Use(middlewares.HeaderXBaseURL)
		e.Use(middlewares.HeaderAPIKey)
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOriginFunc: func(origin string) (bool, error) {
				return true, nil
			},
			AllowHeaders: []string{"Origin", "Content-Length", "Content-Type"},
			AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
			MaxAge:       60 * 60 * 24 * 7,
		}))

		params.OpenAIHandler.InstallForEcho("/v1/openai/query", e)
		params.OpenAIHandler.InstallPlaygroundForEcho("/v1/openai/query", "/v1/openai/", e)

		for _, v := range e.Routes() {
			params.Logger.Debug("registered route", zap.String("method", v.Method), zap.String("path", v.Path))
		}

		server := &http.Server{
			Addr:              params.Config.LingticIo.Gateway.GraphQL.Addr,
			Handler:           e,
			ReadHeaderTimeout: time.Minute,
		}

		params.Lifecycle.Append(fx.Hook{
			OnStop: func(ctx context.Context) error {
				closeCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
				defer cancel()
				if err := server.Shutdown(closeCtx); err != nil && err != http.ErrServerClosed {
					params.Logger.Error("shutdown graphql server failed", zap.Error(err))
					return err
				}

				return nil
			},
		})

		return &Server{
			Server: server,
		}
	}
}

func Run() func(logger *logger.Logger, server *Server) error {
	return func(logger *logger.Logger, server *Server) error {
		logger.Info("starting graphql server...")

		listener, err := net.Listen("tcp", server.Addr)
		if err != nil {
			return fmt.Errorf("failed to listen %s: %v", server.Addr, err)
		}

		go func() {
			err = server.Serve(listener)
			if err != nil && err != http.ErrServerClosed {
				logger.Fatal(err.Error())
			}
		}()

		logger.Info("graphql server listening...", zap.String("addr", server.Addr))

		return nil
	}
}
