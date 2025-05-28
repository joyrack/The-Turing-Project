package llm

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/joyrack/cli-chat/config"
	"github.com/rifaideen/talkative"
)

type ChatSession struct {
	client   *talkative.Client
	messages []talkative.ChatMessage
	model    string
}

func NewChatSession(model string) (*ChatSession, error) {
	slog.Info("Creating new LLM chat session...")
	config, err := config.GetSystemConfig()
	if err != nil {
		return nil, err
	}

	llmClient, err := talkative.New(config.OllamaServerUri)
	if err != nil {
		return nil, fmt.Errorf("error while creating LLM Client: %v", err)
	}

	messages := []talkative.ChatMessage{}
	if config.LlmContext != "" {
		messages = append(messages, talkative.ChatMessage{
			Role:    "system",
			Content: config.LlmContext,
		})
	}

	slog.Info("LLM Chat Session created successfully!", "model", model)
	return &ChatSession{
		client:   llmClient,
		messages: messages,
		model:    model,
	}, nil
}

// Blocking function - Do NOT call on the main thread
func (cs *ChatSession) GetResponse(userMessage string) (string, error) {
	cs.messages = append(cs.messages, talkative.ChatMessage{
		Role:    "user",
		Content: userMessage,
	})

	var stringBuilder strings.Builder

	callback := func(cr string, err error) {
		if err != nil {
			slog.Error("LLM mid-response error", "error", err.Error())
			return
		}

		var response talkative.ChatResponse
		if err := json.Unmarshal([]byte(cr), &response); err != nil {
			slog.Error("error in trying to Unmarshal LLM response", "error", err.Error())
			return
		}

		stringBuilder.WriteString(response.Message.Content)
	}

	// TODO: maybe try passing &ChatParams{Stream: true} - to get response in single shot
	done, err := cs.client.PlainChat(cs.model, callback, nil, cs.messages...)
	if err != nil {
		slog.Error("Error in getting response from LLM", "error", err.Error())
		return "", err
	}

	<-done
	// add the generated response to chat history
	reply := stringBuilder.String()
	cs.messages = append(cs.messages, talkative.ChatMessage{
		Role:    "assistant",
		Content: reply,
	})
	return reply, nil
}
