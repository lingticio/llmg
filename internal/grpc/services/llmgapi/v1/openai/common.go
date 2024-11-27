package openai

import (
	"encoding/json"

	openaiapiv1 "github.com/lingticio/llmg/apis/llmgapi/v1/openai"
	"github.com/samber/lo"
	"github.com/sashabaranov/go-openai"
)

// TODO: split into small functions
func gRPCRequestToOpenAIRequest(req *openaiapiv1.CreateChatCompletionRequest) openai.ChatCompletionRequest {
	request := openai.ChatCompletionRequest{
		Model: req.GetModel(),
		Messages: lo.Map(req.GetMessages(), func(item *openaiapiv1.ChatCompletionMessage, index int) openai.ChatCompletionMessage {
			message := openai.ChatCompletionMessage{}

			switch {
			case item.GetSystemMessage() != nil:
				systemMessage := item.GetSystemMessage()

				if systemMessage.Name != nil {
					message.Name = systemMessage.GetName()
				}

				message.Content = systemMessage.GetContent()
				message.Role = openai.ChatMessageRoleSystem
			case item.GetUserMessage() != nil:
				userMessage := item.GetUserMessage()

				if userMessage.Name != nil {
					message.Name = userMessage.GetName()
				}

				if userMessage.GetContent() != nil && userMessage.GetContent().GetText() != nil {
					message.Content = userMessage.GetContent().GetText().GetContent()
				} else if userMessage.GetContent() != nil && userMessage.GetContent().GetMulti() != nil {
					message.MultiContent = make([]openai.ChatMessagePart, 0)
					for _, part := range userMessage.GetContent().GetMulti().GetParts() {
						openaiPart := openai.ChatMessagePart{}

						switch {
						case part.GetText() != nil:
							openaiPart.Type = openai.ChatMessagePartTypeText

							openaiPart.Text = part.GetText().GetText()
						case part.GetImage() != nil:
							openaiPart.ImageURL = &openai.ChatMessageImageURL{
								URL: part.GetImage().GetImageUrl().GetUrl(),
							}
							if part.GetImage().ImageUrl.Detail != nil {
								openaiPart.ImageURL.Detail = openai.ImageURLDetail(part.GetImage().GetImageUrl().GetDetail())
							} else {
								openaiPart.ImageURL.Detail = openai.ImageURLDetailAuto
							}
						}

						message.MultiContent = append(message.MultiContent, openaiPart)
					}
				}

				message.Role = openai.ChatMessageRoleUser
			case item.GetAssistantMessage() != nil:
				assistantMessage := item.GetAssistantMessage()

				if assistantMessage.Name != nil {
					message.Name = assistantMessage.GetName()
				}
				if assistantMessage.Content != nil {
					message.Content = assistantMessage.GetContent()
				}
				if assistantMessage.ToolCalls != nil {
					message.ToolCalls = make([]openai.ToolCall, 0)

					for _, toolCall := range assistantMessage.GetToolCalls() {
						openaiToolCall := openai.ToolCall{
							ID:   toolCall.GetId(),
							Type: openai.ToolType(toolCall.GetType()),
							Function: openai.FunctionCall{
								Name:      toolCall.GetFunction().GetName(),
								Arguments: toolCall.GetFunction().GetArguments(),
							},
						}

						message.ToolCalls = append(message.ToolCalls, openaiToolCall)
					}
				}

				message.Role = openai.ChatMessageRoleAssistant
			case item.GetToolMessage() != nil:
				toolMessage := item.GetToolMessage()

				message.Role = openai.ChatMessageRoleTool
				message.Content = toolMessage.GetContent()
				message.ToolCallID = toolMessage.GetToolCallId()
			}

			return message
		}),
	}
	if req.MaxTokens != nil {
		request.MaxTokens = int(req.GetMaxTokens())
	}
	if req.Temperature != nil {
		request.Temperature = req.GetTemperature()
	}
	if req.TopP != nil {
		request.TopP = req.GetTopP()
	}
	if req.N != nil {
		request.N = int(req.GetN())
	}
	if req.Stop != nil {
		request.Stop = req.GetStopArray()
	}
	if req.PresencePenalty != nil {
		request.PresencePenalty = req.GetPresencePenalty()
	}
	if req.GetResponseFormat() != nil {
		request.ResponseFormat = &openai.ChatCompletionResponseFormat{}

		switch {
		case req.GetResponseFormat().GetText() != nil:
			request.ResponseFormat.Type = openai.ChatCompletionResponseFormatTypeText
		case req.GetResponseFormat().GetJsonObject() != nil:
			request.ResponseFormat.Type = openai.ChatCompletionResponseFormatTypeJSONObject
		case req.GetResponseFormat().GetJsonSchema() != nil:
			request.ResponseFormat.Type = openai.ChatCompletionResponseFormatTypeJSONSchema
			request.ResponseFormat.JSONSchema = &openai.ChatCompletionResponseFormatJSONSchema{
				Name:        req.GetResponseFormat().GetJsonSchema().GetJsonSchema().GetName(),
				Description: req.GetResponseFormat().GetJsonSchema().GetJsonSchema().GetDescription(),
				Schema:      json.RawMessage([]byte(req.GetResponseFormat().GetJsonSchema().GetJsonSchema().GetSchema())),
				Strict:      req.GetResponseFormat().GetJsonSchema().GetJsonSchema().GetStrict(),
			}
		}
	}
	if req.Seed != nil {
		request.Seed = lo.ToPtr(int(req.GetSeed()))
	}
	if req.FrequencyPenalty != nil {
		request.FrequencyPenalty = req.GetFrequencyPenalty()
	}
	if req.LogitBias != nil {
		request.LogitBias = lo.FromEntries(
			lo.Map(
				lo.Entries(req.GetLogitBias()),
				func(item lo.Entry[string, int64], index int) lo.Entry[string, int] {
					return lo.Entry[string, int]{
						Key:   item.Key,
						Value: int(item.Value),
					}
				},
			),
		)
	}
	if req.LogProbs != nil {
		request.LogProbs = req.GetLogProbs()
	}
	if req.TopLogProbs != nil {
		request.TopLogProbs = int(req.GetTopLogProbs())
	}
	if req.User != nil {
		request.User = req.GetUser()
	}
	if req.Tools != nil {
		request.Tools = make([]openai.Tool, 0)

		for _, tool := range req.GetTools() {
			openaiTool := openai.Tool{
				Type: openai.ToolType(tool.GetType()),
				Function: &openai.FunctionDefinition{
					Name:        tool.GetFunction().GetName(),
					Description: tool.GetFunction().GetDescription(),
					Parameters:  tool.GetFunction().GetParameters(),
				},
			}

			request.Tools = append(request.Tools, openaiTool)
		}
	}
	if req.GetToolChoice() != nil {
		request.ToolChoice = req.GetToolChoice()
	}
	if req.ParallelToolCalls != nil {
		request.ParallelToolCalls = req.GetParallelToolCalls()
	}

	return request
}

// TODO: split into small functions
func gRPCStreamRequestToOpenAIRequest(req *openaiapiv1.CreateChatCompletionStreamRequest) openai.ChatCompletionRequest {
	request := openai.ChatCompletionRequest{
		Model: req.GetModel(),
		Messages: lo.Map(req.GetMessages(), func(item *openaiapiv1.ChatCompletionMessage, index int) openai.ChatCompletionMessage {
			message := openai.ChatCompletionMessage{}

			switch {
			case item.GetSystemMessage() != nil:
				systemMessage := item.GetSystemMessage()

				if systemMessage.Name != nil {
					message.Name = systemMessage.GetName()
				}

				message.Content = systemMessage.GetContent()
				message.Role = openai.ChatMessageRoleSystem
			case item.GetUserMessage() != nil:
				userMessage := item.GetUserMessage()

				if userMessage.Name != nil {
					message.Name = userMessage.GetName()
				}

				if userMessage.GetContent() != nil && userMessage.GetContent().GetText() != nil {
					message.Content = userMessage.GetContent().GetText().GetContent()
				} else if userMessage.GetContent() != nil && userMessage.GetContent().GetMulti() != nil {
					message.MultiContent = make([]openai.ChatMessagePart, 0)
					for _, part := range userMessage.GetContent().GetMulti().GetParts() {
						openaiPart := openai.ChatMessagePart{}

						switch {
						case part.GetText() != nil:
							openaiPart.Type = openai.ChatMessagePartTypeText

							openaiPart.Text = part.GetText().GetText()
						case part.GetImage() != nil:
							openaiPart.ImageURL = &openai.ChatMessageImageURL{
								URL: part.GetImage().GetImageUrl().GetUrl(),
							}
							if part.GetImage().ImageUrl.Detail != nil {
								openaiPart.ImageURL.Detail = openai.ImageURLDetail(part.GetImage().GetImageUrl().GetDetail())
							} else {
								openaiPart.ImageURL.Detail = openai.ImageURLDetailAuto
							}
						}

						message.MultiContent = append(message.MultiContent, openaiPart)
					}
				}

				message.Role = openai.ChatMessageRoleUser
			case item.GetAssistantMessage() != nil:
				assistantMessage := item.GetAssistantMessage()

				if assistantMessage.Name != nil {
					message.Name = assistantMessage.GetName()
				}
				if assistantMessage.Content != nil {
					message.Content = assistantMessage.GetContent()
				}
				if assistantMessage.ToolCalls != nil {
					message.ToolCalls = make([]openai.ToolCall, 0)

					for _, toolCall := range assistantMessage.GetToolCalls() {
						openaiToolCall := openai.ToolCall{
							ID:   toolCall.GetId(),
							Type: openai.ToolType(toolCall.GetType()),
							Function: openai.FunctionCall{
								Name:      toolCall.GetFunction().GetName(),
								Arguments: toolCall.GetFunction().GetArguments(),
							},
						}

						message.ToolCalls = append(message.ToolCalls, openaiToolCall)
					}
				}

				message.Role = openai.ChatMessageRoleAssistant
			case item.GetToolMessage() != nil:
				toolMessage := item.GetToolMessage()

				message.Role = openai.ChatMessageRoleTool
				message.Content = toolMessage.GetContent()
				message.ToolCallID = toolMessage.GetToolCallId()
			}

			return message
		}),
		Stream: true,
	}
	if req.MaxTokens != nil {
		request.MaxTokens = int(req.GetMaxTokens())
	}
	if req.Temperature != nil {
		request.Temperature = req.GetTemperature()
	}
	if req.TopP != nil {
		request.TopP = req.GetTopP()
	}
	if req.N != nil {
		request.N = int(req.GetN())
	}
	if req.Stop != nil {
		request.Stop = req.GetStopArray()
	}
	if req.PresencePenalty != nil {
		request.PresencePenalty = req.GetPresencePenalty()
	}
	if req.GetResponseFormat() != nil {
		request.ResponseFormat = &openai.ChatCompletionResponseFormat{}

		switch {
		case req.GetResponseFormat().GetText() != nil:
			request.ResponseFormat.Type = openai.ChatCompletionResponseFormatTypeText
		case req.GetResponseFormat().GetJsonObject() != nil:
			request.ResponseFormat.Type = openai.ChatCompletionResponseFormatTypeJSONObject
		case req.GetResponseFormat().GetJsonSchema() != nil:
			request.ResponseFormat.Type = openai.ChatCompletionResponseFormatTypeJSONSchema
			request.ResponseFormat.JSONSchema = &openai.ChatCompletionResponseFormatJSONSchema{
				Name:        req.GetResponseFormat().GetJsonSchema().GetJsonSchema().GetName(),
				Description: req.GetResponseFormat().GetJsonSchema().GetJsonSchema().GetDescription(),
				Schema:      json.RawMessage([]byte(req.GetResponseFormat().GetJsonSchema().GetJsonSchema().GetSchema())),
				Strict:      req.GetResponseFormat().GetJsonSchema().GetJsonSchema().GetStrict(),
			}
		}
	}
	if req.Seed != nil {
		request.Seed = lo.ToPtr(int(req.GetSeed()))
	}
	if req.FrequencyPenalty != nil {
		request.FrequencyPenalty = req.GetFrequencyPenalty()
	}
	if req.LogitBias != nil {
		request.LogitBias = lo.FromEntries(
			lo.Map(
				lo.Entries(req.GetLogitBias()),
				func(item lo.Entry[string, int64], index int) lo.Entry[string, int] {
					return lo.Entry[string, int]{
						Key:   item.Key,
						Value: int(item.Value),
					}
				},
			),
		)
	}
	if req.LogProbs != nil {
		request.LogProbs = req.GetLogProbs()
	}
	if req.TopLogProbs != nil {
		request.TopLogProbs = int(req.GetTopLogProbs())
	}
	if req.User != nil {
		request.User = req.GetUser()
	}
	if req.Tools != nil {
		request.Tools = make([]openai.Tool, 0)

		for _, tool := range req.GetTools() {
			openaiTool := openai.Tool{
				Type: openai.ToolType(tool.GetType()),
				Function: &openai.FunctionDefinition{
					Name:        tool.GetFunction().GetName(),
					Description: tool.GetFunction().GetDescription(),
					Parameters:  tool.GetFunction().GetParameters(),
				},
			}

			request.Tools = append(request.Tools, openaiTool)
		}
	}
	if req.GetToolChoice() != nil {
		request.ToolChoice = req.GetToolChoice()
	}
	if req.ParallelToolCalls != nil {
		request.ParallelToolCalls = req.GetParallelToolCalls()
	}
	if req.GetStreamOptions() != nil {
		request.StreamOptions = &openai.StreamOptions{
			IncludeUsage: req.GetStreamOptions().GetIncludeUsage(),
		}
	}

	return request
}

func multiContentToParts(multiContent []openai.ChatMessagePart) []*openaiapiv1.ChatCompletionMessageContentPart {
	return lo.Map(multiContent, func(item openai.ChatMessagePart, index int) *openaiapiv1.ChatCompletionMessageContentPart {
		switch item.Type {
		case openai.ChatMessagePartTypeText:
			return &openaiapiv1.ChatCompletionMessageContentPart{
				Type: &openaiapiv1.ChatCompletionMessageContentPart_Text{
					Text: &openaiapiv1.ChatCompletionMessageContentPartText{
						Text: item.Text,
					},
				},
			}
		case openai.ChatMessagePartTypeImageURL:
			if item.ImageURL == nil {
				return &openaiapiv1.ChatCompletionMessageContentPart{
					Type: &openaiapiv1.ChatCompletionMessageContentPart_Text{
						Text: &openaiapiv1.ChatCompletionMessageContentPartText{
							Text: "",
						},
					},
				}
			}

			return &openaiapiv1.ChatCompletionMessageContentPart{
				Type: &openaiapiv1.ChatCompletionMessageContentPart_Image{
					Image: &openaiapiv1.ChatCompletionMessageContentPartImage{
						ImageUrl: &openaiapiv1.ChatCompletionMessageContentPartImageURL{
							Url:    item.ImageURL.URL,
							Detail: lo.ToPtr(mapOpenAIImageDetailToChatCompletionMessageContentPartImageDetail[item.ImageURL.Detail]),
						},
					},
				},
			}
		default:
			return &openaiapiv1.ChatCompletionMessageContentPart{
				Type: &openaiapiv1.ChatCompletionMessageContentPart_Text{
					Text: &openaiapiv1.ChatCompletionMessageContentPartText{
						Text: "",
					},
				},
			}
		}
	})
}

func logProbsToTokenLogProbs(logProbs []openai.LogProb) []*openaiapiv1.ChatCompletionTokenLogProb {
	return lo.Map(logProbs, func(item openai.LogProb, index int) *openaiapiv1.ChatCompletionTokenLogProb {
		return &openaiapiv1.ChatCompletionTokenLogProb{
			Token:   item.Token,
			LogProb: item.LogProb,
			Bytes:   item.Bytes,
			TopLogProbs: lo.Map(item.TopLogProbs, func(item openai.TopLogProbs, index int) *openaiapiv1.ChatCompletionTokenLogprobTopLogProb {
				return &openaiapiv1.ChatCompletionTokenLogprobTopLogProb{
					Token:   item.Token,
					LogProb: item.LogProb,
					Bytes:   item.Bytes,
				}
			}),
		}
	})
}

func mapMessage(message openai.ChatCompletionMessage) *openaiapiv1.ChatCompletionMessage {
	switch message.Role {
	case openai.ChatMessageRoleSystem:
		systemMessage := &openaiapiv1.ChatCompletionSystemMessage{
			Role:    message.Role,
			Content: message.Content,
			Name:    lo.ToPtr(message.Name),
		}

		return &openaiapiv1.ChatCompletionMessage{
			Message: &openaiapiv1.ChatCompletionMessage_SystemMessage{
				SystemMessage: systemMessage,
			},
		}
	case openai.ChatMessageRoleAssistant:
		assistantMessage := &openaiapiv1.ChatCompletionAssistantMessage{
			Role: message.Role,
			Name: lo.ToPtr(message.Name),
		}
		if message.Content != "" {
			assistantMessage.Content = lo.ToPtr(message.Content)
		}
		if message.ToolCalls != nil {
			assistantMessage.ToolCalls = lo.Map(message.ToolCalls, func(item openai.ToolCall, index int) *openaiapiv1.ChatCompletionMessageToolCall {
				toolCall := openaiapiv1.ChatCompletionMessageToolCall{
					Id: item.ID,
					Function: &openaiapiv1.ChatCompletionMessageToolCallFunction{
						Name:      item.Function.Name,
						Arguments: item.Function.Arguments,
					},
					Type: mapOpenAIToolTypeToChatCompletionMessageToolCallType[item.Type],
				}

				return &toolCall
			})
		}

		return &openaiapiv1.ChatCompletionMessage{
			Message: &openaiapiv1.ChatCompletionMessage_AssistantMessage{
				AssistantMessage: assistantMessage,
			},
		}
	case openai.ChatMessageRoleUser:
		userMessage := &openaiapiv1.ChatCompletionUserMessage{
			Role: message.Role,
			Name: lo.ToPtr(message.Name),
		}
		if message.Content != "" {
			userMessage.Content = &openaiapiv1.ChatCompletionUserMessageContent{
				Content: &openaiapiv1.ChatCompletionUserMessageContent_Text{
					Text: &openaiapiv1.ChatCompletionMessageTextContent{
						Content: message.Content,
					},
				},
			}
		}
		if len(message.MultiContent) > 0 {
			userMessage.Content = &openaiapiv1.ChatCompletionUserMessageContent{
				Content: &openaiapiv1.ChatCompletionUserMessageContent_Multi{
					Multi: &openaiapiv1.ChatCompletionMessageMultiContent{
						Parts: multiContentToParts(message.MultiContent),
					},
				},
			}
		}

		return &openaiapiv1.ChatCompletionMessage{
			Message: &openaiapiv1.ChatCompletionMessage_UserMessage{
				UserMessage: userMessage,
			},
		}
	case openai.ChatMessageRoleTool:
		toolMessage := &openaiapiv1.ChatCompletionToolMessage{
			ToolCallId: message.ToolCallID,
		}

		return &openaiapiv1.ChatCompletionMessage{
			Message: &openaiapiv1.ChatCompletionMessage_ToolMessage{
				ToolMessage: toolMessage,
			},
		}
	default:
		return &openaiapiv1.ChatCompletionMessage{}
	}
}

func mapDelta(message openai.ChatCompletionStreamChoiceDelta) *openaiapiv1.ChatCompletionChunkChoiceDelta {
	delta := &openaiapiv1.ChatCompletionChunkChoiceDelta{
		Role: lo.ToPtr(message.Role),
	}
	if message.Content != "" {
		delta.Content = lo.ToPtr(message.Content)
	}
	if message.ToolCalls != nil {
		delta.ToolCalls = lo.Map(message.ToolCalls, func(item openai.ToolCall, index int) *openaiapiv1.ChatCompletionChunkDeltaToolCall {
			toolCall := openaiapiv1.ChatCompletionChunkDeltaToolCall{
				Index: int64(lo.FromPtr(item.Index)),
				Id:    lo.ToPtr(item.ID),
				Type:  lo.ToPtr(mapOpenAIToolTypeToChatCompletionMessageToolCallType[item.Type]),
				Function: &openaiapiv1.ChatCompletionMessageToolCallFunction{
					Name:      item.Function.Name,
					Arguments: item.Function.Arguments,
				},
			}

			return &toolCall
		})
	}

	return delta
}

var mapOpenAIToolTypeToChatCompletionMessageToolCallType = map[openai.ToolType]openaiapiv1.ChatCompletionMessageToolCallType{
	openai.ToolTypeFunction: openaiapiv1.ChatCompletionMessageToolCallType_ChatCompletionMessageToolCallTypeFunction,
}

var mapOpenAIImageDetailToChatCompletionMessageContentPartImageDetail = map[openai.ImageURLDetail]openaiapiv1.ChatCompletionMessageContentPartImageDetail{
	openai.ImageURLDetailAuto: openaiapiv1.ChatCompletionMessageContentPartImageDetail_ChatCompletionMessageContentPartImageDetailAuto,
	openai.ImageURLDetailHigh: openaiapiv1.ChatCompletionMessageContentPartImageDetail_ChatCompletionMessageContentPartImageDetailHigh,
	openai.ImageURLDetailLow:  openaiapiv1.ChatCompletionMessageContentPartImageDetail_ChatCompletionMessageContentPartImageDetailLow,
}

var mapOpenAIFinishedReasonToChatCompletionFinishReason = map[openai.FinishReason]openaiapiv1.ChatCompletionFinishReason{
	openai.FinishReasonStop:          openaiapiv1.ChatCompletionFinishReason_ChatCompletionFinishReasonStop,
	openai.FinishReasonLength:        openaiapiv1.ChatCompletionFinishReason_ChatCompletionFinishReasonLength,
	openai.FinishReasonFunctionCall:  openaiapiv1.ChatCompletionFinishReason_ChatCompletionFinishReasonFunctionCall,
	openai.FinishReasonContentFilter: openaiapiv1.ChatCompletionFinishReason_ChatCompletionFinishReasonContentFilter,
	openai.FinishReasonNull:          openaiapiv1.ChatCompletionFinishReason_ChatCompletionFinishReasonNull,
}
