package servers

import (
	"github.com/lingticio/llmg/internal/grpc/servers/apiserver"
	v1 "github.com/lingticio/llmg/internal/grpc/servers/llmg/v1"
	"go.uber.org/fx"
)

func Modules() fx.Option {
	return fx.Options(
		fx.Provide(apiserver.NewGRPCServer()),
		fx.Provide(apiserver.NewGatewayServer()),
		fx.Provide(v1.NewV1GRPCServer()),
	)
}
