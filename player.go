package main

type Role int

const (
	QUESTIONER Role = iota
	ANSWERER
)

func (role Role) String() string {
	if role == QUESTIONER {
		return "Questioner"
	} else {
		return "Answerer"
	}
}

type Player struct {
	id       string
	username string
	role     Role
}
