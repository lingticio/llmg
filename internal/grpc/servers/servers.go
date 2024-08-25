package servers

import (
	"github.com/lingticio/gateway/internal/grpc/servers/apiserver"
	"go.uber.org/fx"
)

func Modules() fx.Option {
	return fx.Options(
		fx.Provide(apiserver.NewGRPCServer()),
		fx.Provide(apiserver.NewGatewayServer()),
	)
}
