package configs

import (
	"github.com/samber/lo"
	"github.com/spf13/viper"
)

type LingticIoGatewayAPIServerHttpServer struct {
	Addr string `json:"server_addr" yaml:"server_addr"`
}

type LingticIoGatewayAPIServerGrpcServer struct {
	Addr string `json:"server_addr" yaml:"server_addr"`
}

type LingticIoGatewayAPIServerGraphQLServer struct {
	Addr string `json:"server_addr" yaml:"server_addr"`
}

type LingticIoGateway struct {
	Http    LingticIoGatewayAPIServerHttpServer    `json:"http" yaml:"http"`
	Grpc    LingticIoGatewayAPIServerGrpcServer    `json:"grpc" yaml:"grpc"`
	GraphQL LingticIoGatewayAPIServerGraphQLServer `json:"graphql" yaml:"graphql"`
}

type LingticIo struct {
	Gateway LingticIoGateway `json:"core" yaml:"core"`
}

func defaultLingticIoConfig() LingticIo {
	return LingticIo{
		LingticIoGateway{
			Http: LingticIoGatewayAPIServerHttpServer{
				Addr: ":8080",
			},
			Grpc: LingticIoGatewayAPIServerGrpcServer{
				Addr: ":8081",
			},
			GraphQL: LingticIoGatewayAPIServerGraphQLServer{
				Addr: ":8082",
			},
		},
	}
}

func registerLingticIoCoreConfig() {
	lo.Must0(viper.BindEnv("lingticio.core.http.server_addr"))
	lo.Must0(viper.BindEnv("lingticio.core.grpc.server_addr"))
	lo.Must0(viper.BindEnv("lingticio.core.graphql.server_addr"))
}
