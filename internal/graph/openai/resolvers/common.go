package resolvers

import (
	"encoding/json"

	"github.com/lingticio/llmg/internal/graph/openai/model"
	"github.com/nekomeowww/fo"
	"github.com/samber/lo"
	"github.com/sashabaranov/go-openai"
)

func inputToRequest(input model.CreateChatCompletionInput, stream bool) openai.ChatCompletionRequest {
	request := openai.ChatCompletionRequest{
		Model:  input.Model,
		Stream: stream,
		Messages: lo.Map(input.Messages, func(item *model.ChatCompletionMessageInput, index int) openai.ChatCompletionMessage {
			message := openai.ChatCompletionMessage{
				Role: item.Role,
			}

			if item.Content != nil {
				message.Content = *item.Content
			} else if item.MultiContent != nil {
				message.MultiContent = make([]openai.ChatMessagePart, 0)
				for _, part := range item.MultiContent {
					openaiPart := openai.ChatMessagePart{
						Type: openai.ChatMessagePartType(part.Type),
					}

					switch part.Type {
					case string(openai.ChatMessagePartTypeText):
						if part.Text != nil {
							openaiPart.Text = *part.Text
						}
					case string(openai.ChatMessagePartTypeImageURL):
						openaiPart.ImageURL = &openai.ChatMessageImageURL{
							URL: part.ImageURL.URL,
						}
						if part.ImageURL.Detail != nil {
							openaiPart.ImageURL.Detail = openai.ImageURLDetail(*part.ImageURL.Detail)
						} else {
							openaiPart.ImageURL.Detail = openai.ImageURLDetailAuto
						}
					}

					message.MultiContent = append(message.MultiContent, openaiPart)
				}
			}

			if item.Name != nil {
				message.Name = *item.Name
			}
			if item.ToolCalls != nil {
				message.ToolCalls = make([]openai.ToolCall, 0)

				for _, toolCall := range item.ToolCalls {
					openaiToolCall := openai.ToolCall{
						ID:   toolCall.ID,
						Type: openai.ToolType(toolCall.Type),
						Function: openai.FunctionCall{
							Name:      toolCall.Function.Name,
							Arguments: toolCall.Function.Arguments,
						},
					}

					message.ToolCalls = append(message.ToolCalls, openaiToolCall)
				}
			}
			if item.ToolCallID != nil {
				message.ToolCallID = *item.ToolCallID
			}

			return message
		}),
	}
	if input.MaxTokens != nil {
		request.MaxTokens = *input.MaxTokens
	}
	if input.Temperature != nil {
		request.Temperature = float32(*input.Temperature)
	}
	if input.TopP != nil {
		request.TopP = float32(*input.TopP)
	}
	if input.N != nil {
		request.N = *input.N
	}
	if input.Stop != nil {
		request.Stop = lo.Map(lo.Filter(input.Stop, func(item *string, index int) bool {
			return item != nil
		}), func(item *string, index int) string {
			return *item
		})
	}
	if input.PresencePenalty != nil {
		request.PresencePenalty = float32(*input.PresencePenalty)
	}
	if input.ResponseFormat != nil {
		request.ResponseFormat = &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatType(input.ResponseFormat.Type),
		}
		if input.ResponseFormat.JSONSchema != nil {
			request.ResponseFormat.JSONSchema = &openai.ChatCompletionResponseFormatJSONSchema{
				Name:        input.ResponseFormat.JSONSchema.Name,
				Description: input.ResponseFormat.JSONSchema.Description,
			}
			if input.ResponseFormat.JSONSchema.Schema != nil {
				request.ResponseFormat.JSONSchema.Schema = json.RawMessage(fo.May(json.Marshal(input.ResponseFormat.JSONSchema.Schema)))
			}
			if input.ResponseFormat.JSONSchema.Strict != nil {
				request.ResponseFormat.JSONSchema.Strict = *input.ResponseFormat.JSONSchema.Strict
			}
		}
	}
	if input.Seed != nil {
		request.Seed = input.Seed
	}
	if input.FrequencyPenalty != nil {
		request.FrequencyPenalty = float32(*input.FrequencyPenalty)
	}
	if input.LogitBias != nil {
		request.LogitBias = lo.FromEntries(
			lo.Map(
				lo.Entries(input.LogitBias),
				func(item lo.Entry[string, any], index int) lo.Entry[string, int] {
					value, ok := item.Value.(int)
					if !ok {
						return lo.Entry[string, int]{
							Key: item.Key,
						}
					}

					return lo.Entry[string, int]{
						Key:   item.Key,
						Value: value,
					}
				},
			),
		)
	}
	if input.LogProbs != nil {
		request.LogProbs = *input.LogProbs
	}
	if input.TopLogProbs != nil {
		request.TopLogProbs = *input.TopLogProbs
	}
	if input.User != nil {
		request.User = *input.User
	}
	if input.Tools != nil {
		request.Tools = make([]openai.Tool, 0)

		for _, tool := range input.Tools {
			openaiTool := openai.Tool{
				Type: openai.ToolType(tool.Type),
				Function: &openai.FunctionDefinition{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					Parameters:  tool.Function.Parameters,
				},
			}

			request.Tools = append(request.Tools, openaiTool)
		}
	}
	if input.ToolChoice != nil {
		request.ToolChoice = *input.ToolChoice
	}
	if input.StreamOptions != nil {
		request.StreamOptions = &openai.StreamOptions{}
		if input.StreamOptions.IncludeUsage != nil {
			request.StreamOptions.IncludeUsage = *input.StreamOptions.IncludeUsage
		}
	}
	if input.ParallelToolCalls != nil {
		request.ParallelToolCalls = *input.ParallelToolCalls
	}

	return request
}

func multiContentToParts(multiContent []openai.ChatMessagePart) []model.ChatCompletionMessageContentPart {
	return lo.Map(multiContent, func(item openai.ChatMessagePart, index int) model.ChatCompletionMessageContentPart {
		switch item.Type {
		case openai.ChatMessagePartTypeText:
			return model.ChatCompletionContentPartText{
				Text: item.Text,
			}
		case openai.ChatMessagePartTypeImageURL:
			if item.ImageURL == nil {
				return model.ChatCompletionContentPartText{}
			}

			return model.ChatCompletionContentPartImage{
				ImageURL: &model.ChatCompletionContentPartImageURL{
					URL:    item.ImageURL.URL,
					Detail: lo.ToPtr(model.ImageDetail(item.ImageURL.Detail)),
				},
			}
		default:
			return model.ChatCompletionContentPartText{}
		}
	})
}

func logProbsToTokenLogProbs(logProbs []openai.LogProb) []*model.TokenLogProb {
	return lo.Map(logProbs, func(item openai.LogProb, index int) *model.TokenLogProb {
		return &model.TokenLogProb{
			Token:   item.Token,
			LogProb: item.LogProb,
			Bytes: lo.Map(item.Bytes, func(item byte, index int) int {
				return index
			}),
			TopLogProbs: lo.Map(item.TopLogProbs, func(item openai.TopLogProbs, index int) *model.TopLogProb {
				return &model.TopLogProb{
					Token:   item.Token,
					LogProb: item.LogProb,
					Bytes: lo.Map(item.Bytes, func(item byte, index int) int {
						return index
					}),
				}
			}),
		}
	})
}

func mapMessage(message openai.ChatCompletionMessage) model.ChatCompletionMessage {
	switch message.Role {
	case openai.ChatMessageRoleSystem:
		systemMessage := model.ChatCompletionSystemMessage{
			Role: message.Role,
			Content: model.ChatCompletionTextContent{
				Text: message.Content,
			},
			Name: lo.ToPtr(message.Name),
		}

		return systemMessage
	case openai.ChatMessageRoleAssistant:
		assistantMessage := model.ChatCompletionAssistantMessage{
			Role:       message.Role,
			Name:       lo.ToPtr(message.Name),
			ToolCallID: lo.ToPtr(message.ToolCallID),
		}
		if message.Content != "" {
			assistantMessage.Content = model.ChatCompletionTextContent{
				Text: message.Content,
			}
		}
		if len(message.MultiContent) > 0 {
			assistantMessage.Content = model.ChatCompletionArrayContent{
				Parts: multiContentToParts(message.MultiContent),
			}
		}
		if message.ToolCalls != nil {
			assistantMessage.ToolCalls = lo.Map(message.ToolCalls, func(item openai.ToolCall, index int) *model.ChatCompletionMessageToolCall {
				toolCall := model.ChatCompletionMessageToolCall{
					ID: item.ID,
					Function: &model.FunctionCall{
						Name:      item.Function.Name,
						Arguments: item.Function.Arguments,
					},
					Type: string(item.Type),
				}

				return &toolCall
			})
		}

		return assistantMessage
	case openai.ChatMessageRoleUser:
		userMessage := model.ChatCompletionUserMessage{
			Role: message.Role,
			Name: lo.ToPtr(message.Name),
		}
		if message.Content != "" {
			userMessage.Content = model.ChatCompletionTextContent{
				Text: message.Content,
			}
		}
		if len(message.MultiContent) > 0 {
			userMessage.Content = model.ChatCompletionArrayContent{
				Parts: multiContentToParts(message.MultiContent),
			}
		}

		return userMessage
	default:
		return model.ChatCompletionUserMessage{}
	}
}

func mapDelta(message openai.ChatCompletionStreamChoiceDelta) *model.ChatCompletionStreamResponseDelta {
	delta := &model.ChatCompletionStreamResponseDelta{
		Role: message.Role,
	}
	if message.Content != "" {
		delta.Content = lo.ToPtr(message.Content)
	}
	if message.FunctionCall != nil {
		delta.FunctionCall = &model.FunctionCall{
			Name:      message.FunctionCall.Name,
			Arguments: message.FunctionCall.Arguments,
		}
	}
	if message.ToolCalls != nil {
		delta.ToolCalls = lo.Map(message.ToolCalls, func(item openai.ToolCall, index int) *model.ChatCompletionMessageToolCallChunk {
			toolCall := model.ChatCompletionMessageToolCallChunk{
				Index: item.Index,
				ID:    item.ID,
				Type:  string(item.Type),
				Function: &model.FunctionCallChunk{
					Name:      item.Function.Name,
					Arguments: item.Function.Arguments,
				},
			}

			return &toolCall
		})
	}

	return delta
}
