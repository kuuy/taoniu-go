package binance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

type ServerRepository struct {
	Rdb *redis.Client
	Ctx context.Context
}

type ServerTime struct {
	Timestamp int64 `json:"serverTime"`
}

func (r *ServerRepository) Time() (int64, error) {
	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	session := &net.Dialer{}
	tr.DialContext = session.DialContext

	httpClient := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(100) * time.Millisecond,
	}

	timestamp := time.Now().UnixNano() / int64(time.Millisecond)

	url := "https://api.binance.com/api/v1/time"
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, errors.New(
			fmt.Sprintf(
				"request error: status[%s] code[%d]",
				resp.Status,
				resp.StatusCode,
			),
		)
	}

	var result ServerTime
	json.NewDecoder(resp.Body).Decode(&result)

	r.Rdb.HSet(
		r.Ctx,
		"binance:server",
		map[string]interface{}{
			"timestamp": result.Timestamp,
			"timediff":  timestamp - result.Timestamp,
		},
	)

	return result.Timestamp, nil
}
