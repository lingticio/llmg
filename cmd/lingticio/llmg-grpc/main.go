package main

import (
	"context"
	"log"
	"time"

	"go.uber.org/fx"

	"github.com/lingticio/llmg/internal/configs"
	"github.com/lingticio/llmg/internal/datastore"
	grpcservers "github.com/lingticio/llmg/internal/grpc/servers"
	v1 "github.com/lingticio/llmg/internal/grpc/servers/llmg/v1"
	grpcservices "github.com/lingticio/llmg/internal/grpc/services"
	"github.com/lingticio/llmg/internal/libs"
	"github.com/spf13/cobra"
)

var (
	configFilePath string
	envFilePath    string
)

func main() {
	root := &cobra.Command{
		Use: "llmg-grpc",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := fx.New(
				fx.Provide(configs.NewConfig("lingticio", "llmg", configFilePath, envFilePath)),
				fx.Options(libs.Modules()),
				fx.Options(datastore.Modules()),
				fx.Options(grpcservers.Modules()),
				fx.Options(grpcservices.Modules()),
				fx.Invoke(v1.Run()),
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
