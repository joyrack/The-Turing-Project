package controllers

import (
	"github.com/joyrack/cli-chat/models"
	"github.com/rivo/tview"
)

// methods that the controller needs from the view
type PlayerView interface {
	App() *tview.Application
	StartGame()
	UpdateMessageView(msg *models.Message)
}
