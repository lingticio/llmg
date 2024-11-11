package configs

import (
	"github.com/lingticio/llmg/internal/meta"
	"github.com/lingticio/llmg/pkg/types/metadata"
	"github.com/samber/lo"
	"github.com/spf13/viper"
)

type HttpServer struct {
	Addr string `json:"server_addr" yaml:"server_addr"`
}

type GrpcServer struct {
	Addr string `json:"server_addr" yaml:"server_addr"`
}

type GraphQLServer struct {
	Addr string `json:"server_addr" yaml:"server_addr"`
}

type Endpoint struct {
	ID       string                             `json:"id" yaml:"id"`
	Alias    string                             `json:"alias" yaml:"alias"`
	APIKey   string                             `json:"api_key" yaml:"api_key"`
	Upstream *metadata.UpstreamSingleOrMultiple `json:"upstream,omitempty" yaml:"upstream,omitempty"`
}

type Group struct {
	ID string `json:"id" yaml:"id"`

	Groups    []Group                            `json:"groups" yaml:"groups"`
	Endpoints []Endpoint                         `json:"endpoints" yaml:"endpoints"`
	Upstream  *metadata.UpstreamSingleOrMultiple `json:"upstream,omitempty" yaml:"upstream,omitempty"`
}

type Team struct {
	ID string `json:"id" yaml:"id"`

	Groups   []Group                            `json:"groups" yaml:"groups"`
	Upstream *metadata.UpstreamSingleOrMultiple `json:"upstream,omitempty" yaml:"upstream,omitempty"`
}

type Tenant struct {
	ID string `json:"id" yaml:"id"`

	Teams    []Team                             `json:"teams" yaml:"teams"`
	Upstream *metadata.UpstreamSingleOrMultiple `json:"upstream,omitempty" yaml:"upstream,omitempty"`
}

type Routes struct {
	Tenants  []Tenant                           `json:"tenants" yaml:"tenants"`
	Upstream *metadata.UpstreamSingleOrMultiple `json:"upstream,omitempty" yaml:"upstream,omitempty"`
}

type Config struct {
	meta.Meta `json:"-" yaml:"-"`

	Env string `json:"env" yaml:"env"`

	Http    HttpServer    `json:"http" yaml:"http"`
	Grpc    GrpcServer    `json:"grpc" yaml:"grpc"`
	GraphQL GraphQLServer `json:"graphql" yaml:"graphql"`
	Routes  Routes        `json:"configs" yaml:"configs"`
}

func defaultConfig() Config {
	return Config{
		Http: HttpServer{
			Addr: ":8080",
		},
		Grpc: GrpcServer{
			Addr: ":8081",
		},
		GraphQL: GraphQLServer{
			Addr: ":8082",
		},
	}
}

func registerLingticIoCoreConfig() {
	lo.Must0(viper.BindEnv("lingticio.llmg.http.server_addr"))
	lo.Must0(viper.BindEnv("lingticio.llmg.grpc.server_addr"))
	lo.Must0(viper.BindEnv("lingticio.llmg.graphql.server_addr"))
}
