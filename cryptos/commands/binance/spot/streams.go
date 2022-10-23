package spot

import (
	"context"
	"fmt"
	"github.com/gammazero/workerpool"
	"gorm.io/gorm"
	"strconv"
	"strings"
	models "taoniu.local/cryptos/models/binance"
	"time"

	"nhooyr.io/websocket"

	"github.com/bitly/go-simplejson"
	"github.com/go-redis/redis/v8"
	"github.com/urfave/cli/v2"

	pool "taoniu.local/cryptos/common"
)

type StreamsHandler struct {
	ID      int64
	Db      *gorm.DB
	Rdb     *redis.Client
	Ctx     context.Context
	Symbols []string
}

func NewStreamCommand() *cli.Command {
	h := StreamsHandler{
		ID:      1,
		Db:      pool.NewDB(),
		Rdb:     pool.NewRedis(),
		Ctx:     context.Background(),
		Symbols: []string{},
	}

	return &cli.Command{
		Name:  "streams",
		Usage: "",
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
	redisKey := fmt.Sprintf("binance:spot:realtime:%s", data["s"])
	value, err := h.Rdb.HGet(h.Ctx, redisKey, "price").Result()
	if err != nil {
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
	h.online()
	defer h.offline()

	streams := []string{}
	for _, symbol := range h.Symbols {
		streams = append(
			streams,
			fmt.Sprintf("%s@miniTicker", strings.ToLower(symbol)),
		)
	}
	if len(streams) < 1 {
		return nil
	}
	endpoint := "wss://stream.binance.com/stream?streams=" + strings.Join(streams, "/")

	wp := workerpool.New(30)
	defer wp.StopWait()

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
		wp.Submit(func() {
			h.handler(message)
		})
	}

	return nil
}

func (h *StreamsHandler) online() error {
	var symbols []string

	symbols, _ = h.Rdb.ZRangeByScore(
		h.Ctx,
		"binance:spot:streams:symbols",
		&redis.ZRangeBy{
			Min: fmt.Sprintf("%d", h.ID),
			Max: fmt.Sprintf("%d", h.ID),
		},
	).Result()
	for _, symbol := range symbols {
		h.Rdb.ZRem(
			h.Ctx,
			"binance:spot:streams:symbols",
			symbol,
		).Result()
	}

	h.Db.Model(models.Symbol{}).Select("symbol").Where("status=? AND is_spot=True", "TRADING").Find(&symbols)
	for _, symbol := range symbols {
		score, _ := h.Rdb.ZScore(
			h.Ctx,
			"binance:spot:streams:symbols",
			symbol,
		).Result()
		if score > 0 {
			continue
		}
		h.append(symbol)
		if len(h.Symbols) >= 30 {
			break
		}
	}

	return nil
}

func (h *StreamsHandler) append(symbol string) error {
	mutex := pool.NewMutex(
		h.Rdb,
		h.Ctx,
		fmt.Sprintf("locks:binance:spot:streams:symbols:%s", symbol),
	)
	if mutex.Lock(5 * time.Second) {
		return nil
	}
	defer mutex.Unlock()

	h.Rdb.ZAdd(
		h.Ctx,
		"binance:spot:streams:symbols",
		&redis.Z{Score: float64(h.ID), Member: symbol},
	).Result()
	h.Symbols = append(h.Symbols, symbol)

	return nil
}

func (h *StreamsHandler) offline() error {
	for _, symbol := range h.Symbols {
		h.Rdb.ZRem(
			h.Ctx,
			"binance:spot:streams:symbols",
			symbol,
		).Result()
	}

	return nil
}
