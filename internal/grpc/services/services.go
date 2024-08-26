package services

import (
	"github.com/lingticio/llmg/internal/grpc/services/llmgapi/v1/openai"
	"go.uber.org/fx"
)

func Modules() fx.Option {
	return fx.Options(
		fx.Provide(openai.NewOpenAIService()),
	)
}
