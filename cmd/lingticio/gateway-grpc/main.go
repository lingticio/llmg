package main

import (
	"context"
	"log"
	"time"

	"go.uber.org/fx"

	"github.com/lingticio/gateway/internal/configs"
	"github.com/lingticio/gateway/internal/datastore"
	grpcservers "github.com/lingticio/gateway/internal/grpc/servers"
	apiserver "github.com/lingticio/gateway/internal/grpc/servers/apiserver"
	grpcservices "github.com/lingticio/gateway/internal/grpc/services"
	"github.com/lingticio/gateway/internal/libs"
	"github.com/spf13/cobra"
)

var (
	configFilePath string
	envFilePath    string
)

func main() {
	root := &cobra.Command{
		Use: "gateway-grpc",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := fx.New(
				fx.Provide(configs.NewConfig("lingticio", "gateway", configFilePath, envFilePath)),
				fx.Options(libs.Modules()),
				fx.Options(datastore.Modules()),
				fx.Options(grpcservers.Modules()),
				fx.Options(grpcservices.Modules()),
				fx.Invoke(apiserver.RunGRPCServer()),
				fx.Invoke(apiserver.RunGatewayServer()),
			)

			app.Run()

			stopCtx, stopCtxCancel := context.WithTimeout(context.Background(), time.Minute*5)
			defer stopCtxCancel()

			if err := app.Stop(stopCtx); err != nil {
				return err
			}

			return nil
		},
	}

	root.Flags().StringVarP(&configFilePath, "config", "c", "", "config file path")
	root.Flags().StringVarP(&envFilePath, "env", "e", "", "env file path")

	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}
