package gatewayapi

import (
	"go.uber.org/fx"
	"google.golang.org/grpc/reflection"

	gatewayapiv1 "github.com/lingticio/llmg/apis/gatewayapi/v1"
	serverv1 "github.com/lingticio/llmg/internal/grpc/services/gatewayapi/v1/server"
	grpcpkg "github.com/lingticio/llmg/pkg/grpc"
)

func Modules() fx.Option {
	return fx.Options(
		fx.Provide(NewCoreAPI()),
		fx.Provide(serverv1.NewServerService()),
	)
}

type NewCoreAPIParams struct {
	fx.In

	Server *serverv1.ServerService
}

type CoreAPI struct {
	params *NewCoreAPIParams
}

func NewCoreAPI() func(params NewCoreAPIParams) *CoreAPI {
	return func(params NewCoreAPIParams) *CoreAPI {
		return &CoreAPI{params: &params}
	}
}

func (c *CoreAPI) Register(r *grpcpkg.Register) {
	r.RegisterHttpHandlers([]grpcpkg.HttpHandler{
		gatewayapiv1.RegisterServerServiceHandler,
	})
	r.RegisterGrpcService(func(s reflection.GRPCServer) {
		gatewayapiv1.RegisterServerServiceServer(s, c.params.Server)
	})
}
