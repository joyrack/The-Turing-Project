// This file contains the code for establishing a peer to peer connection b/w 2 clients
package main

import (
	"github.com/redis/go-redis/v9"
)

type Connection struct {
	questionerId string
	answererId   string
	Id           string
}

func getOpponentRole(p *Player) Role {
	if p.role == QUESTIONER {
		return ANSWERER
	} else {
		return QUESTIONER
	}
}

func EstablishConnection(player *Player, rdb *redis.Client) (*Connection, error) {
	// ------------ MULTI --------------
	_, err := rdb.RPush(ctx, player.role.String(), player.id).Result()
	if err != nil {
		return nil, err
	}

	opponentId, err := rdb.LPop(ctx, getOpponentRole(player).String()).Result()
	// ----------- EXEC -----------------
	if err == redis.Nil {
		res, _ := rdb.BLPop(ctx, 0, getOpponentRole(player).String()).Result()

		if len(res) == 1 {
			opponentId = res[0]
		}
	} else if err != nil {
		return nil, err
	}

	var conn *Connection
	if player.role == QUESTIONER {
		conn = &Connection{questionerId: player.id, answererId: opponentId}
	} else {
		conn = &Connection{questionerId: opponentId, answererId: player.id}
	}
	return conn, nil
}
