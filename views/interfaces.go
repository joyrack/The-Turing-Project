package views

import "github.com/joyrack/cli-chat/models"

// methods that the view needs from the controller
type PlayerController interface {
	SetPlayerRole(models.Role)
	AddMessage(models.Message)
}
