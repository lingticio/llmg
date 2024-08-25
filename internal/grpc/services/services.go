package services

import (
	"go.uber.org/fx"
	"google.golang.org/grpc/reflection"

	"github.com/lingticio/gateway/internal/grpc/services/gatewayapi"
	grpcpkg "github.com/lingticio/gateway/pkg/grpc"
)

func Modules() fx.Option {
	return fx.Options(
		fx.Provide(NewRegister()),
		fx.Options(gatewayapi.Modules()),
	)
}

type NewRegisterParams struct {
	fx.In

	GatewayAPI *gatewayapi.CoreAPI
}

func NewRegister() func(params NewRegisterParams) *grpcpkg.Register {
	return func(params NewRegisterParams) *grpcpkg.Register {
		register := grpcpkg.NewRegister()

		params.GatewayAPI.Register(register)

		register.RegisterGrpcService(func(s reflection.GRPCServer) {
			reflection.Register(s)
		})

		return register
	}
}
