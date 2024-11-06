package configs

import (
	"github.com/lingticio/llmg/internal/meta"
	"github.com/lingticio/llmg/pkg/util/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

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

func NewConfig(namespace string, app string, configFilePath string, envFilePath string) func() (*Config, error) {
	return func() (*Config, error) {
		var configPath string
		if utils.IsInUnitTest() {
			configPath = tryToMatchConfigPathForUnitTest(configFilePath)
		} else {
			configPath = getConfigFilePath(configFilePath)
		}

		registerLingticIoCoreConfig()

		err := loadEnvConfig(envFilePath)
		if err != nil {
			return nil, err
		}

		err = readConfig(configPath)
		if err != nil {
			return nil, err
		}

		config := defaultConfig()

		err = viper.Unmarshal(&config, func(c *mapstructure.DecoderConfig) {
			c.TagName = "yaml"
		})
		if err != nil {
			return nil, err
		}

		meta.Env = config.Env

		config.Meta.Env = config.Env
		config.Meta.App = app
		config.Meta.Namespace = namespace

		return &config, nil
	}
}
