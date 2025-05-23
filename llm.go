package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rifaideen/talkative"
)

var llmClient *talkative.Client

type Llm struct {
	messageBroker *MessageBroker
}

// Blocking function - Do NOT call on the Main thread
func (llm *Llm) GetResponse(query string) {
	message := talkative.ChatMessage{
		Role:    talkative.USER,
		Content: query,
	}

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
		// llm.messageBroker.SendMessage("LLM", response.Message.Content)
	}

	done, err := llmClient.PlainChat("tinyllama", callback, nil, message)
	if err != nil {
		panic(fmt.Sprintf("Chat error: %v", err))
	}

	<-done
	llm.messageBroker.SendMessage("LLM", stringBuilder.String())
}

func init() {
	var err error
	llmClient, err = talkative.New("http://localhost:11434")
	if err != nil {
		panic(fmt.Sprintf("failed to create talkative client: %v", err))
	} else {
		fmt.Println("LLM client created successfully!")
	}

}
