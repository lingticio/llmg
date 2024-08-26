package openai

import (
	"context"

	openaiapiv1 "github.com/lingticio/llmg/apis/llmgapi/v1/openai"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OpenAIService struct {
	openaiapiv1.UnimplementedOpenAIServiceServer
}

func NewOpenAIService() func() *OpenAIService {
	return func() *OpenAIService {
		return &OpenAIService{}
	}
}

func (s *OpenAIService) CreateChatCompletion(ctx context.Context, req *openaiapiv1.CreateChatCompletionRequest) (*openaiapiv1.CreateChatCompletionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateChatCompletion not implemented")
}
func (s *OpenAIService) CreateChatCompletionStream(req *openaiapiv1.CreateChatCompletionStreamRequest, server openaiapiv1.OpenAIService_CreateChatCompletionStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method CreateChatCompletionStream not implemented")
}
