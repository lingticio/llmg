package configs

import (
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
	ID     string `json:"id" yaml:"id"`
	Alias  string `json:"alias" yaml:"alias"`
	APIKey string `json:"api_key" yaml:"api_key"`
}

type Group struct {
	ID string `json:"id" yaml:"id"`

	Groups    []Group    `json:"groups" yaml:"groups"`
	Endpoints []Endpoint `json:"endpoints" yaml:"endpoints"`
}

type Team struct {
	ID string `json:"id" yaml:"id"`

	Groups []Group `json:"groups" yaml:"groups"`
}

type Tenant struct {
	ID string `json:"id" yaml:"id"`

	Teams []Team `json:"teams" yaml:"teams"`
}

type Configs struct {
	Tenants []Tenant `json:"tenants" yaml:"tenants"`
}

func registerLingticIoCoreConfig() {
	lo.Must0(viper.BindEnv("lingticio.llmg.http.server_addr"))
	lo.Must0(viper.BindEnv("lingticio.llmg.grpc.server_addr"))
	lo.Must0(viper.BindEnv("lingticio.llmg.graphql.server_addr"))
}
