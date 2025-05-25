package controllers

import (
	"github.com/joyrack/cli-chat/models"
)

type PlayerController struct {
	playerModel *models.PlayerModel
	playerView  PlayerView
}
