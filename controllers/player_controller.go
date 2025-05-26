package controllers

import (
	"log/slog"

	"github.com/joyrack/cli-chat/models"
)

type PlayerController struct {
	model  *models.PlayerModel
	view   PlayerView
	player models.Player
	conn   *models.Connection
}

func NewPlayerController(model *models.PlayerModel, view PlayerView) *PlayerController {
	return &PlayerController{
		model: model,
		view:  view,
	}
}

// Blocking function
func (controller *PlayerController) InitializeGame(username string, role models.Role) {
	// 1. assign this player a unique ID
	id, err := controller.model.GetNextPlayerId()
	if err != nil {
		// what to do?
		slog.Error("Error in initializing player", "error", err.Error())
		return
	}
	slog.Info("Player initialized successfully", "id", id, "username", username)

	controller.player = models.Player{
		Id:       id,
		Username: username,
		Role:     role,
	}

	conn, err := controller.model.FindOpponent(&controller.player)
	if err != nil {
		slog.Error("Error in finding opponent", "error", err.Error())
		// what to do?
		return
		// maybe something like view.showError() ??
	}
	connId, err := conn.Id()
	if err != nil {
		slog.Error("Could not establish connection", "error", err.Error())
		return
	}
	slog.Info("Connection established successfully", "connection_id", connId)
	controller.conn = conn
	controller.view.App().QueueUpdateDraw(func() {
		controller.view.StartGame()
	})
}

func (controller *PlayerController) AddMessage(msg *models.Message) {
	_, err := controller.model.AddToStream(controller.conn, msg)
	if err != nil {
		// view.showError()
	}

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
