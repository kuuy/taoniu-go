package spot

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"nhooyr.io/websocket"

	"github.com/bitly/go-simplejson"
	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"

	pool "taoniu.local/cryptos/common"
)

type StreamsHandler struct {
	Rdb *redis.Client
	Ctx context.Context
}

func NewStreamCommand() *cli.Command {
	h := StreamsHandler{
		Rdb: pool.NewRedis(),
		Ctx: context.Background(),
	}

	return &cli.Command{
		Name:  "streams",
		Usage: "",
		Action: func(c *cli.Context) error {
			if err := h.start(); err != nil {
				return cli.Exit(err.Error(), 1)
			}
			return nil
		},
	}
}

func (h *StreamsHandler) newJSON(data []byte) (j *simplejson.Json, err error) {
	j, err = simplejson.NewJson(data)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (h *StreamsHandler) handler(message []byte) {
	j, err := h.newJSON(message)
	if err != nil {
		panic(err)
	}

	data := j.Get("data").MustMap()

	if data["e"] != "24hrMiniTicker" {
		return
	}

	timestamp := time.Now().Unix()
	redisKey := fmt.Sprintf("binance:spot:realtime:%s", data["s"])
	value, err := h.Rdb.HGet(h.Ctx, redisKey, "price").Result()
	if err != redis.Nil {
		lasttime, _ := strconv.ParseInt(value, 10, 64)
		if lasttime > timestamp {
			return
		}
	}
	h.Rdb.HMSet(
		h.Ctx,
		redisKey,
		map[string]interface{}{
			"symbol":    data["s"],
			"price":     data["c"],
			"open":      data["o"],
			"high":      data["h"],
			"low":       data["l"],
			"volume":    data["v"],
			"quota":     data["q"],
			"timestamp": fmt.Sprint(timestamp),
		},
	)
}

func (h *StreamsHandler) start() error {
	symbols, _ := h.Rdb.SMembers(h.Ctx, "binance:spot:websocket:symbols").Result()
	streams := []string{}
	for _, symbol := range symbols {
		streams = append(
			streams,
			fmt.Sprintf("%s@miniTicker", strings.ToLower(symbol)),
		)
	}
	endpoint := "wss://stream.binance.com/stream?streams=" + strings.Join(streams, "/")

	socket, _, err := websocket.Dial(h.Ctx, endpoint, nil)
	if err != nil {
		return err
	}
	socket.SetReadLimit(655350)

	for {
		_, message, readErr := socket.Read(h.Ctx)
		if readErr != nil {
			return readErr
		}
		h.handler(message)
	}

	return nil
}
