package configs

import (
	"github.com/samber/lo"
	"github.com/spf13/viper"
)

type LingticIoLLMGHttpServer struct {
	Addr string `json:"server_addr" yaml:"server_addr"`
}

type LingticIoLLMGGrpcServer struct {
	Addr string `json:"server_addr" yaml:"server_addr"`
}

type LingticIoLLMGGraphQLServer struct {
	Addr string `json:"server_addr" yaml:"server_addr"`
}

type LingticIoLLMG struct {
	Http    LingticIoLLMGHttpServer    `json:"http" yaml:"http"`
	Grpc    LingticIoLLMGGrpcServer    `json:"grpc" yaml:"grpc"`
	GraphQL LingticIoLLMGGraphQLServer `json:"graphql" yaml:"graphql"`
}

type LingticIo struct {
	LLMG LingticIoLLMG `json:"core" yaml:"core"`
}

func defaultLingticIoConfig() LingticIo {
	return LingticIo{
		LingticIoLLMG{
			Http: LingticIoLLMGHttpServer{
				Addr: ":8080",
			},
			Grpc: LingticIoLLMGGrpcServer{
				Addr: ":8081",
			},
			GraphQL: LingticIoLLMGGraphQLServer{
				Addr: ":8082",
			},
		},
	}
}

func registerLingticIoCoreConfig() {
	lo.Must0(viper.BindEnv("lingticio.llmg.http.server_addr"))
	lo.Must0(viper.BindEnv("lingticio.llmg.grpc.server_addr"))
	lo.Must0(viper.BindEnv("lingticio.llmg.graphql.server_addr"))
}
