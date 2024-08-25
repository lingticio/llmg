package servers

import (
	"github.com/lingticio/llmg/internal/grpc/servers/apiserver"
	"go.uber.org/fx"
)

func Modules() fx.Option {
	return fx.Options(
		fx.Provide(apiserver.NewGRPCServer()),
		fx.Provide(apiserver.NewGatewayServer()),
	)
}
