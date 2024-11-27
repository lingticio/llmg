package openai

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/nekomeowww/xo/logger"
	"github.com/samber/lo"
	"github.com/sashabaranov/go-openai"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	openaiapiv1 "github.com/lingticio/llmg/apis/llmgapi/v1/openai"
)

type NewOpenAIServiceParams struct {
	fx.In

	Logger *logger.Logger
}

type OpenAIService struct {
	openaiapiv1.UnimplementedOpenAIServiceServer

	logger *logger.Logger
}

func NewOpenAIService() func(params NewOpenAIServiceParams) *OpenAIService {
	return func(params NewOpenAIServiceParams) *OpenAIService {
		return &OpenAIService{
			logger: params.Logger,
		}
	}
}

func clientConfigFromContext(ctx context.Context) (openai.ClientConfig, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return openai.ClientConfig{}, status.Errorf(codes.Internal, "failed to get metadata")
	}

	apiKeys := md.Get("x-api-key")
	if len(apiKeys) == 0 {
		return openai.ClientConfig{}, status.Errorf(codes.InvalidArgument, "missing API key in x-api-key")
	}

	config := openai.DefaultConfig(apiKeys[0])

	baseURLs := md.Get("x-base-url")
	if len(baseURLs) > 0 {
		config.BaseURL = baseURLs[0]
	}

	return config, nil
}

func (s *OpenAIService) CreateChatCompletion(ctx context.Context, req *openaiapiv1.CreateChatCompletionRequest) (*openaiapiv1.CreateChatCompletionResponse, error) {
	config, err := clientConfigFromContext(ctx)
	if err != nil {
		return nil, err
	}

	client := openai.NewClientWithConfig(config)

	openaiResponse, err := client.CreateChatCompletion(ctx, gRPCRequestToOpenAIRequest(req))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create chat completion: %v", err)
	}

	response := &openaiapiv1.CreateChatCompletionResponse{
		Id:      openaiResponse.ID,
		Object:  openaiResponse.Object,
		Created: timestamppb.New(time.Unix(openaiResponse.Created, 0)),
		Model:   openaiResponse.Model,
		Choices: lo.Map(openaiResponse.Choices, func(item openai.ChatCompletionChoice, index int) *openaiapiv1.ChatCompletionChoice {
			choice := &openaiapiv1.ChatCompletionChoice{
				Index:        int64(item.Index),
				Message:      mapMessage(item.Message),
				FinishReason: openaiapiv1.ChatCompletionFinishReason(openaiapiv1.ChatCompletionFinishReason_value[string(item.FinishReason)]),
			}
			if item.LogProbs != nil {
				choice.LogProbs = new(openaiapiv1.ChatCompletionChoiceLogProbs)
			}
			if item.LogProbs != nil && item.LogProbs.Content != nil {
				choice.LogProbs.Content = logProbsToTokenLogProbs(item.LogProbs.Content)
			}

			return choice
		}),
		ServiceTier:       lo.ToPtr(req.GetServiceTier()),
		SystemFingerprint: lo.ToPtr(openaiResponse.SystemFingerprint),
		Usage: &openaiapiv1.ChatCompletionUsage{
			PromptTokens:     int64(openaiResponse.Usage.PromptTokens),
			CompletionTokens: int64(openaiResponse.Usage.CompletionTokens),
			TotalTokens:      int64(openaiResponse.Usage.TotalTokens),
		},
	}

	return response, nil
}

func (s *OpenAIService) CreateChatCompletionStream(req *openaiapiv1.CreateChatCompletionStreamRequest, server openaiapiv1.OpenAIService_CreateChatCompletionStreamServer) error {
	config, err := clientConfigFromContext(server.Context())
	if err != nil {
		return err
	}

	client := openai.NewClientWithConfig(config)

	stream, err := client.CreateChatCompletionStream(server.Context(), gRPCStreamRequestToOpenAIRequest(req))
	if err != nil {
		return status.Errorf(codes.Internal, "failed to create chat completion stream: %v", err)
	}

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			s.logger.Info("stream closed")
			break
		}
		if err != nil {
			s.logger.Error("failed to receive chat completion stream", zap.Error(err))
			return status.Errorf(codes.Internal, "failed to receive chat completion stream: %v", err)
		}

		chunkResponse := &openaiapiv1.CreateChatCompletionStreamResponse{
			Id:      response.ID,
			Object:  response.Object,
			Created: timestamppb.New(time.Unix(response.Created, 0)),
			Model:   response.Model,
			Choices: lo.Map(response.Choices, func(item openai.ChatCompletionStreamChoice, index int) *openaiapiv1.ChatCompletionChunkChoice {
				choice := &openaiapiv1.ChatCompletionChunkChoice{
					Index:        int64(item.Index),
					Delta:        mapDelta(item.Delta),
					FinishReason: lo.ToPtr(mapOpenAIFinishedReasonToChatCompletionFinishReason[item.FinishReason]),
				}

				return choice
			}),
			SystemFingerprint: lo.ToPtr(response.SystemFingerprint),
		}
		if response.Usage != nil {
			chunkResponse.Usage = &openaiapiv1.ChatCompletionUsage{
				PromptTokens:     int64(response.Usage.PromptTokens),
				CompletionTokens: int64(response.Usage.CompletionTokens),
				TotalTokens:      int64(response.Usage.TotalTokens),
			}
		}

		if err := server.Send(chunkResponse); err != nil {
			s.logger.Error("failed to send chat completion stream", zap.Error(err))
		}
	}

	return nil
}
