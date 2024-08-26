package v1

import (
	"context"
	"net"
	"net/http"

	openaiapiv1 "github.com/lingticio/llmg/apis/llmgapi/v1/openai"
	"github.com/lingticio/llmg/internal/configs"
	"github.com/lingticio/llmg/internal/grpc/services/llmgapi/v1/openai"
	"github.com/nekomeowww/xo/logger"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type NewV1GRPCServerParam struct {
	fx.In

	Lifecycle     fx.Lifecycle
	Config        *configs.Config
	Logger        *logger.Logger
	OpenAIService *openai.OpenAIService
}

type V1GRPCServer struct {
	GRPCServer *grpc.Server
	Addr       string
}

func NewV1GRPCServer() func(params NewV1GRPCServerParam) *V1GRPCServer {
	return func(params NewV1GRPCServerParam) *V1GRPCServer {
		grpcServer := grpc.NewServer()
		openaiapiv1.RegisterOpenAIServiceServer(grpcServer, params.OpenAIService)
		reflection.Register(grpcServer)

		params.Lifecycle.Append(fx.Hook{
			OnStop: func(ctx context.Context) error {
				params.Logger.Info("gracefully shutting down v1 gRPC server...")
				grpcServer.GracefulStop()
				return nil
			},
		})

		return &V1GRPCServer{
			GRPCServer: grpcServer,
			Addr:       params.Config.LingticIo.LLMG.Grpc.Addr,
		}
	}
}

func Run() func(*logger.Logger, *V1GRPCServer) error {
	return func(logger *logger.Logger, server *V1GRPCServer) error {
		logger.Info("starting v1 gRPC service...")

		listener, err := net.Listen("tcp", server.Addr)
		if err != nil {
			return err
		}

		go func() {
			err := server.GRPCServer.Serve(listener)
			if err != nil && err != http.ErrServerClosed {
				logger.Error("failed to serve v1 gRPC server", zap.Error(err))
			}
		}()

		logger.Info("v1 gRPC server listening", zap.String("addr", listener.Addr().String()))

		return nil
	}
}
