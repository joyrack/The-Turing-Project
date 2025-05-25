package controllers

import (
	"github.com/joyrack/cli-chat/models"
)

type PlayerController struct {
	model  *models.PlayerModel
	view   *PlayerView
	player models.Player
	conn   *models.Connection
}

func (controller *PlayerController) InitializeGame(username string, role models.Role) {
	// 1. assign this player a unique ID
	id, err := controller.model.GetNextPlayerId()
	if err != nil {
		// what to do?
	}

	controller.player = models.Player{
		Id:       id,
		Username: username,
		Role:     role,
	}

	conn, err := controller.model.FindOpponent(&controller.player)
	if err != nil {
		// what to do?
		// maybe something like view.showError() ??
	}
	controller.conn = conn
}

func (controller *PlayerController) AddMessage(msg *models.Message) {
	_, err := controller.model.AddToStream(controller.conn, msg)
	if err != nil {
		// view.showError()
	}

}
