package main

import (
	"context"
	"log"
	"time"

	"go.uber.org/fx"

	"github.com/lingticio/llmg/internal/configs"
	"github.com/lingticio/llmg/internal/datastore"
	"github.com/lingticio/llmg/internal/graph/server"
	"github.com/lingticio/llmg/internal/libs"
	"github.com/spf13/cobra"
)

var (
	configFilePath string
	envFilePath    string
)

func main() {
	root := &cobra.Command{
		Use: "llmg-graphql",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := fx.New(
				fx.Provide(configs.NewConfig("lingticio", "llmg", configFilePath, envFilePath)),
				fx.Options(libs.Modules()),
				fx.Options(datastore.Modules()),
				fx.Options(server.Modules()),
				fx.Invoke(server.Run()),
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
