package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rifaideen/talkative"
)

type ChatSession struct {
	client        *talkative.Client
	messages      []talkative.ChatMessage
	model         string
	messageBroker *MessageBroker
}

func NewChatSession(model string, messageBroker *MessageBroker) (*ChatSession, error) {
	config, err := GetSystemConfig()
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

	return &ChatSession{
		client:        llmClient,
		messages:      messages,
		model:         model,
		messageBroker: messageBroker,
	}, nil
}

// Blocking function - Do NOT call on the main thread
func (cs *ChatSession) GenerateResponse(userMessage string) {
	cs.messages = append(cs.messages, talkative.ChatMessage{
		Role:    "user",
		Content: userMessage,
	})

	var stringBuilder strings.Builder

	callback := func(cr string, err error) {
		if err != nil {
			fmt.Println(err)
			return
		}

		var response talkative.ChatResponse
		if err := json.Unmarshal([]byte(cr), &response); err != nil {
			fmt.Println(err)
			return
		}

		stringBuilder.WriteString(response.Message.Content)
	}

	config, ercallbackr := GetSystemConfig()
	if ercallbackr != nil {
		panic(fmt.Sprintf("Error getting system config: %v", ercallbackr))
	}
	fmt.Println("uri:", config.OllamaServerUri)
	// TODO: maybe try passing &ChatParams{Stream: true} - to get response in single shot
	done, err := cs.client.PlainChat(cs.model, callback, nil, cs.messages...)
	if err != nil {
		panic(fmt.Sprintf("Chat error: %v", err))
	}

	<-done
	// add the generated response to chat history
	reply := stringBuilder.String()
	cs.messages = append(cs.messages, talkative.ChatMessage{
		Role:    "assistant",
		Content: reply,
	})
	cs.messageBroker.SendMessage("LLM", reply)
}
