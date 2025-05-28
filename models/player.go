package models

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

type Role int
type GameMode int

const (
	QUESTIONER Role = iota
	ANSWERER
)

const (
	HUMAN GameMode = iota
	MACHINE
)

func (role Role) String() string {
	if role == QUESTIONER {
		return "Questioner"
	} else {
		return "Answerer"
	}
}

func (role Role) Other() Role {
	if role == QUESTIONER {
		return ANSWERER
	} else {
		return QUESTIONER
	}
}

func (mode GameMode) String() string {
	if mode == HUMAN {
		return "Human"
	} else {
		return "Machine"
	}
}

// data model
type Player struct {
	Id       string
	Username string
	Role     Role
}

func (player *Player) IsValid() (bool, error) {
	_, err := strconv.Atoi(player.Id)
	if err != nil {
		return false, fmt.Errorf("invalid player id: %s", player.Id)
	}

	if player.Role != QUESTIONER && player.Role != ANSWERER {
		return false, fmt.Errorf("invalid player role")
	}

	if player.Username == "" {
		return false, fmt.Errorf("invalid player username. username cannot be empty")
	}
	return true, nil
}

type Message struct {
	Sender  string
	Content string
	id      string
}

func (msg *Message) Id() string {
	return msg.id
}

type Connection struct {
	questionerId string
	answererId   string
	id           string
}

func (conn *Connection) Id() string {
	return conn.id
}

// Handles all database operations
type PlayerModel struct {
	client *redis.Client
	ctx    context.Context
}

func NewPlayerModel(client *redis.Client) *PlayerModel {
	return &PlayerModel{
		client: client,
		ctx:    context.Background(),
	}
}

// database specific operations
func (playerModel *PlayerModel) AddToStream(conn *Connection, msg *Message) (string, error) {
	stream := conn.Id()

	id, err := playerModel.client.XAdd(playerModel.ctx, &redis.XAddArgs{
		Stream: stream,
		ID:     "*",
		Values: []interface{}{"sender", msg.Sender, "content", msg.Content},
	}).Result()

	// TODO: Confirm that redis.Nil is not returned by XAdd
	if err != nil {
		return "", fmt.Errorf("unable to add message to stream. %w", err)
	}
	msg.id = id
	return id, nil
}

// Blocking function
func (playerModel *PlayerModel) GetFromStream(conn *Connection, lastMessageId string) (*Message, error) {
	stream := conn.Id()

	res, err := playerModel.client.XRead(playerModel.ctx, &redis.XReadArgs{
		Block:   0,
		Streams: []string{stream},
		Count:   1,
		ID:      lastMessageId,
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("error in reading from stream. %w", err)
	}

	var msg *Message

	for i := range res {
		redisStream := res[i]
		if redisStream.Stream != stream {
			continue
		}

		if len(redisStream.Messages) == 0 {
			return nil, fmt.Errorf("no messages in stream")
		}

		msg, err = convertToMessage(&redisStream.Messages[0])
		if err != nil {
			return nil, fmt.Errorf("error in converting redis.XMessage to Message: %w", err)
		}
	}

	if msg == nil {
		return nil, fmt.Errorf("could not find in specified stream %s", stream)
	}
	return msg, nil
}

func (playerModel *PlayerModel) GetNextPlayerId() (string, error) {
	id, err := playerModel.client.Incr(playerModel.ctx, "ID").Result()
	if err != nil {
		return "", fmt.Errorf("error in fetching next player id: %w", err)
	}

	return strconv.FormatInt(id, 10), nil
}

func (playerModel *PlayerModel) CreateConnection(questionerId string, answererId string) (*Connection, error) {
	if questionerId == "" || answererId == "" {
		return nil, fmt.Errorf("invalid parameters - both questionerId & answererId should be a non-empty string")
	}

	qId, err := strconv.Atoi(questionerId)
	if err != nil {
		return nil, fmt.Errorf("invalid questionerId: %s. error - %w", questionerId, err)
	}

	aId, err := strconv.Atoi(answererId)
	if err != nil {
		return nil, fmt.Errorf("invalid answererId: %s. error - %w", answererId, err)
	}

	var id string
	if qId < aId {
		id = questionerId + ":" + answererId
	} else {
		id = answererId + ":" + questionerId
	}

	return &Connection{questionerId: questionerId, answererId: answererId, id: id}, nil
}

// Blocking function
func (playerModel *PlayerModel) FindOpponent(player *Player) (*Connection, error) {
	if ok, err := player.IsValid(); !ok {
		return nil, fmt.Errorf("error in finding opponent: %w", err)
	}
	slog.Info("Adding to waiting queue", "Queue", player.Role.String(), "Value", player.Id)
	_, err := playerModel.client.LPush(playerModel.ctx, player.Role.String(), player.Id).Result()
	if err != nil {
		return nil, err
	}

	opponentId, err := playerModel.client.RPop(playerModel.ctx, player.Role.Other().String()).Result()
	if err == redis.Nil {
		slog.Info("Waiting for opponent...")
		res, er := playerModel.client.BRPop(playerModel.ctx, 0, player.Role.Other().String()).Result()
		slog.Info("Found", "res", strings.Join(res, ","))
		if er != nil {
			return nil, er
		}

		if len(res) == 0 {
			return nil, fmt.Errorf("could not find any opponent")
		}
		opponentId = res[1]
	} else if err != nil {
		return nil, fmt.Errorf("error in trying to find opponent: %w", err)
	}

	slog.Info("Opponent found", "opponent_id", opponentId)
	var conn *Connection
	if player.Role == QUESTIONER {
		conn, err = playerModel.CreateConnection(player.Id, opponentId)
	} else {
		conn, err = playerModel.CreateConnection(opponentId, player.Id)
	}

	return conn, err
}

func convertToMessage(msg *redis.XMessage) (*Message, error) {
	var sender string
	var content string
	var ok bool

	if sender, ok = msg.Values["sender"].(string); !ok {
		return nil, fmt.Errorf("incorrect message schema. no valid `sender` field present")
	}

	if content, ok = msg.Values["content"].(string); !ok {
		return nil, fmt.Errorf("incorrect message schema. no valid `content` field present")
	}

	if sender == "" {
		return nil, fmt.Errorf("sender field cannot be empty")
	} else if content == "" {
		return nil, fmt.Errorf("content field cannot be empty")
	}

	return &Message{Sender: sender, Content: content, id: msg.ID}, nil
}
