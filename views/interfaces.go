package views

import "github.com/joyrack/cli-chat/models"

// methods that the view needs from the controller
type PlayerController interface {
	InitializeGame(string, models.Role)
	AddMessage(*models.Message)
}
