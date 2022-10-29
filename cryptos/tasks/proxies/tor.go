package proxies

import (
	"context"
	"strconv"
	pool "taoniu.local/cryptos/common"
	"time"

	"github.com/go-redis/redis/v8"

	repositories "taoniu.local/cryptos/repositories/proxies"
)

type TorTask struct {
	Rdb        *redis.Client
	Ctx        context.Context
	Repository *repositories.TorRepository
}

func (t *TorTask) Flush() error {
	mutex := pool.NewMutex(
		t.Rdb,
		t.Ctx,
		"lock:proxies:tor:flush",
	)
	if mutex.Lock(5 * time.Second) {
		return nil
	}
	defer mutex.Unlock()

	items, _ := t.Rdb.SMembers(t.Ctx, "proxies:tor:offline").Result()
	timestamp := time.Now().Unix()
	for _, item := range items {
		port, _ := strconv.Atoi(item)
		err := t.Repository.ChangeIp(port)
		if err != nil {
			continue
		}
		t.Rdb.SRem(t.Ctx, "proxies:tor:offline", port)
		score, _ := t.Rdb.ZScore(t.Ctx, "proxies:tor:checker", strconv.Itoa(port)).Result()
		if score > 0 {
			continue
		}
		t.Rdb.ZAdd(
			t.Ctx,
			"proxies:tor:checker",
			&redis.Z{
				float64(timestamp),
				port,
			},
		)
	}

	return nil
}

func (t *TorTask) Offline() error {
	items, _ := t.Rdb.ZRangeByScore(
		t.Ctx,
		"proxies:tor:failed",
		&redis.ZRangeBy{
			Min: strconv.Itoa(3),
			Max: "+inf",
		},
	).Result()
	for _, item := range items {
		port, _ := strconv.Atoi(item)
		t.Repository.Offline(port)
		t.Rdb.ZRem(t.Ctx, "proxies:tor:failed", port)
	}

	return nil
}

func (t *TorTask) Checker() error {
	mutex := pool.NewMutex(
		t.Rdb,
		t.Ctx,
		"lock:proxies:tor:flush",
	)
	if mutex.Lock(30 * time.Second) {
		return nil
	}
	defer mutex.Unlock()

	timestamp := time.Now().Unix() - 30
	items, _ := t.Rdb.ZRangeByScore(
		t.Ctx,
		"proxies:tor:checker",
		&redis.ZRangeBy{
			Min: "-inf",
			Max: strconv.FormatInt(timestamp, 10),
		},
	).Result()
	for _, item := range items {
		port, _ := strconv.Atoi(item)
		t.Repository.Checker(port)
	}

	return nil
}
