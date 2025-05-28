package controllers

import (
	"fmt"
	"log/slog"
	"math/rand/v2"

	"github.com/joyrack/cli-chat/config"
	"github.com/joyrack/cli-chat/llm"
	"github.com/joyrack/cli-chat/models"
)

type PlayerController struct {
	model          *models.PlayerModel
	view           PlayerView
	player         *models.Player
	conn           *models.Connection
	gameMode       models.GameMode
	llmChatSession *llm.ChatSession
}

func NewPlayerController(model *models.PlayerModel, view PlayerView) *PlayerController {
	return &PlayerController{
		model: model,
		view:  view,
	}
}

// Blocking function
func (controller *PlayerController) InitializeGame(username string, role models.Role) {
	slog.Info("Initializing Game...")
	controller.initializePlayer(username, role)

	if controller.gameMode == models.MACHINE {
		controller.initializeLLM()
	} else {
		conn, err := controller.model.FindOpponent(controller.player)
		if err != nil {
			slog.Error("Error in finding opponent", "error", err.Error())
			// what to do?
			return
			// maybe something like view.showError() ??
		}

		slog.Info("Connection established successfully", "connection_id", conn.Id())
		controller.conn = conn
	}
	controller.view.App().QueueUpdateDraw(func() {
		controller.view.StartGame()
	})
}

func (controller *PlayerController) initializePlayer(username string, role models.Role) error {
	id, err := controller.model.GetNextPlayerId()
	if err != nil {
		return fmt.Errorf("error in initializing player: %w", err)
	}

	controller.player = &models.Player{
		Username: username,
		Role:     role,
		Id:       id,
	}

	controller.initializeGameMode()
	slog.Info("Player initialized successfully", "id", id, "username", username, "role", role.String(), "gameMode", controller.gameMode.String())
	return nil
}

// When the player is playing as a Questioner, there are 2 game modes possible:
//  1. Playing against LLM
//  2. Playing against human
//
// When the player is playing as an Answerer, only 1 game mode is possible:
//  1. Playing against human
//
// This function selects an appropriate mode for the player based on their role
func (controller *PlayerController) initializeGameMode() {
	if controller.player.Role == models.ANSWERER {
		controller.gameMode = models.HUMAN
		return
	}

	randomChoice := rand.IntN(2)
	if randomChoice == 0 {
		controller.gameMode = models.HUMAN
	} else {
		controller.gameMode = models.MACHINE
	}
}

func (controller *PlayerController) initializeLLM() error {
	llmId, err := controller.model.GetNextPlayerId()
	if err != nil {
		return fmt.Errorf("error in initializing LLM: %w", err)
	}

	conn, err := controller.model.CreateConnection(controller.player.Id, llmId)
	if err != nil {
		return fmt.Errorf("error in creating connection: %w", err)
	}
	controller.conn = conn

	config, err := config.GetSystemConfig()
	if err != nil {
		return err
	}
	chatSession, err := llm.NewChatSession(config.Model)
	if err != nil {
		return fmt.Errorf("error in establishing a chat session with LLM: %w", err)
	}
	controller.llmChatSession = chatSession
	return nil
}

func (controller *PlayerController) AddMessage(msg *models.Message) {
	_, err := controller.model.AddToStream(controller.conn, msg)
	if err != nil {
		// view.showError()
		slog.Error("Error while adding message to stream", "error", err.Error())
		return
	}

	if controller.gameMode == models.MACHINE {
		response, err := controller.llmChatSession.GetResponse(msg.Content)
		if err != nil {
			slog.Error("Error in getting response from LLM", "error", err.Error())
			return
		}

		_, err = controller.model.AddToStream(controller.conn, &models.Message{Sender: "LLM", Content: response})
		if err != nil {
			slog.Error("Error while adding LLM's response to stream", "error", err.Error())
		}
	}

}

// Add method to reveal opponent type
func (controller *PlayerController) GetOpponentType() string {
	if controller.gameMode == models.MACHINE {
		return "LLM"
	}
	return "Human"
}

// Blocking method
func (controller *PlayerController) StartListeningForMessages() {
	slog.Info("Listening to messages now")
	id := "0"
	for {
		msg, err := controller.model.GetFromStream(controller.conn, id)
		if err != nil {
			// view.ShowError()
			slog.Error("error in trying to fetch message from stream", "error", err.Error())
		}
		id = msg.Id()
		go controller.view.App().QueueUpdateDraw(func() {
			controller.view.UpdateMessageView(msg)
		})
	}
}
