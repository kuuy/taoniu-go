package futures

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"

	"nhooyr.io/websocket"

	"github.com/bitly/go-simplejson"
	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"

	pool "taoniu.local/cryptos/common"
)

var (
	rdb *redis.Client
	ctx context.Context
)

type StreamsHandler struct {
	ID      int64
	Db      *gorm.DB
	Rdb     *redis.Client
	Ctx     context.Context
	Symbols []string
}

func NewStreamCommand() *cli.Command {
	var h StreamsHandler
	return &cli.Command{
		Name:  "streams",
		Usage: "",
		Before: func(c *cli.Context) error {
			h = StreamsHandler{
				ID:      1,
				Db:      pool.NewDB(),
				Rdb:     pool.NewRedis(),
				Ctx:     context.Background(),
				Symbols: []string{},
			}
			return nil
		},
		Action: func(c *cli.Context) error {
			id, err := strconv.ParseInt(c.Args().Get(0), 10, 64)
			if err == nil {
				h.ID = id
			}
			if h.ID < 1 {
				return nil
			}
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
	redisKey := fmt.Sprintf("binance:futures:realtime:%s", data["s"])
	value, err := rdb.HGet(ctx, redisKey, "price").Result()
	if err != redis.Nil {
		lasttime, _ := strconv.ParseInt(value, 10, 64)
		if lasttime >= timestamp {
			return
		}
	}
	rdb.HMSet(
		ctx,
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
	symbols, _ := rdb.SMembers(ctx, "binance:futures:websocket:symbols").Result()
	streams := []string{}
	for _, symbol := range symbols {
		streams = append(
			streams,
			fmt.Sprintf("%s@miniTicker", strings.ToLower(symbol)),
		)
	}
	endpoint := "wss://fstream.binance.com/stream?streams=" + strings.Join(streams, "/")

	socket, _, err := websocket.Dial(ctx, endpoint, nil)
	if err != nil {
		return err
	}
	socket.SetReadLimit(655350)

	for {
		_, message, readErr := socket.Read(ctx)
		if readErr != nil {
			return readErr
		}
		h.handler(message)
	}

	return nil
}
