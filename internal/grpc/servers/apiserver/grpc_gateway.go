package apiserver

import (
	"context"
	"encoding/gob"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nekomeowww/xo/logger"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/lingticio/llmg/internal/configs"
	"github.com/lingticio/llmg/internal/grpc/servers/interceptors"
	"github.com/lingticio/llmg/internal/grpc/servers/middlewares"
	grpcpkg "github.com/lingticio/llmg/pkg/util/grpc"
)

type NewGatewayServerParams struct {
	fx.In

	Lifecycle fx.Lifecycle
	Config    *configs.Config
	Register  *grpcpkg.Register
	Logger    *logger.Logger
}

type GatewayServer struct {
	ListenAddr     string
	GRPCServerAddr string

	echo   *echo.Echo
	server *http.Server
}

func NewGatewayServer() func(params NewGatewayServerParams) (*GatewayServer, error) {
	return func(params NewGatewayServerParams) (*GatewayServer, error) {
		gob.Register(map[interface{}]interface{}{})

		e := echo.New()

		e.Use(middlewares.ResponseLog(params.Logger))
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{
				"http://localhost:3000",
			},
			AllowHeaders: []string{
				echo.HeaderOrigin,
				echo.HeaderContentType,
				echo.HeaderAccept,
				echo.HeaderAuthorization,
			},
		}))
		e.RouteNotFound("/*", middlewares.NotFound)

		for path, methodHandlers := range params.Register.EchoHandlers {
			for method, handler := range methodHandlers {
				e.Add(method, path, handler)
			}
		}

		server := &GatewayServer{
			ListenAddr:     params.Config.Http.Addr,
			GRPCServerAddr: params.Config.Grpc.Addr,
			echo:           e,
			server: &http.Server{
				Addr:              params.Config.Http.Addr,
				Handler:           e,
				ReadHeaderTimeout: time.Duration(30) * time.Second,
			},
		}

		params.Lifecycle.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				conn, err := grpc.NewClient(params.Config.Grpc.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					return err
				}

				gateway, err := grpcpkg.NewGateway(ctx, conn, params.Logger,
					grpcpkg.WithServerMuxOptions(
						runtime.WithErrorHandler(interceptors.HttpErrorHandler(params.Logger)),
						runtime.WithMetadata(interceptors.MetadataCookie()),
						runtime.WithMetadata(interceptors.MetadataAuthorization()),
						runtime.WithMetadata(interceptors.MetadataRequestPath()),
					),
					grpcpkg.WithHandlers(params.Register.HttpHandlers...),
				)
				if err != nil {
					return err
				}

				server.echo.Any("/api/v1/*", echo.WrapHandler(gateway))

				return nil
			},
		})

		return server, nil
	}
}

func RunGatewayServer() func(logger *logger.Logger, server *GatewayServer) error {
	return func(logger *logger.Logger, server *GatewayServer) error {
		logger.Info("starting http server...")

		listener, err := net.Listen("tcp", server.ListenAddr)
		if err != nil {
			return fmt.Errorf("failed to listen %s: %v", server.ListenAddr, err)
		}

		go func() {
			err = server.server.Serve(listener)
			if err != nil && err != http.ErrServerClosed {
				logger.Fatal(err.Error())
			}
		}()

		logger.Info("http server listening...", zap.String("addr", server.ListenAddr))

		return nil
	}
}
