package models

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

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

func (role Role) Other() Role {
	if role == QUESTIONER {
		return ANSWERER
	} else {
		return QUESTIONER
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

type Connection struct {
	questionerId string
	answererId   string
	id           string
}

func (conn *Connection) Id() (string, error) {
	if conn.id != "" {
		return conn.id, nil
	}

	if conn.questionerId == "" || conn.answererId == "" {
		return "", fmt.Errorf("could not create get connection ID due to incomplete connection data")
	}

	questionerId, err := strconv.Atoi(conn.questionerId)
	if err != nil {
		return "", fmt.Errorf("invalid connection: QuestionerId = %s. %v", conn.questionerId, err)
	}
	answererId, err := strconv.Atoi(conn.answererId)
	if err != nil {
		return "", fmt.Errorf("invalid connection: AnsweredId = %s. %v", conn.answererId, err)
	}

	if questionerId == answererId {
		return "", fmt.Errorf("invalid connection. Both QuestionerId & AnswererId equal to %d", questionerId)
	}

	if questionerId < answererId {
		conn.id = conn.questionerId + ":" + conn.answererId
	} else {
		conn.id = conn.answererId + ":" + conn.questionerId
	}
	return conn.id, nil
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
	stream, err := conn.Id()
	if err != nil {
		return "", fmt.Errorf("unable to add message to stream. %w", err)
	}

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
	stream, err := conn.Id()
	if err != nil {
		return nil, fmt.Errorf("unable to read from stream. %w", err)
	}

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

// Blocking function
func (playerModel *PlayerModel) FindOpponent(player *Player) (*Connection, error) {
	if ok, err := player.IsValid(); !ok {
		return nil, fmt.Errorf("error in finding opponent: %w", err)
	}
	_, err := playerModel.client.LPush(playerModel.ctx, player.Role.String(), player.Id).Result()
	if err != nil {
		return nil, err
	}

	opponentId, err := playerModel.client.RPop(playerModel.ctx, player.Role.Other().String()).Result()
	if err == redis.Nil {
		res, er := playerModel.client.BRPop(playerModel.ctx, 0, player.Role.Other().String()).Result()
		if er != nil {
			return nil, er
		}

		if len(res) == 0 {
			return nil, fmt.Errorf("could not find any opponent")
		}
		opponentId = res[0]
	} else if err != nil {
		return nil, fmt.Errorf("error in trying to find opponent: %w", err)
	}

	var conn *Connection
	if player.Role == QUESTIONER {
		conn = &Connection{questionerId: player.Id, answererId: opponentId}
	} else {
		conn = &Connection{questionerId: opponentId, answererId: player.Id}
	}
	return conn, nil
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

	return &Message{Sender: sender, Content: content}, nil
}
