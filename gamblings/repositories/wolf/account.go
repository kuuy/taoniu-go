package wolf

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"

	"taoniu.local/gamblings/common"
	config "taoniu.local/gamblings/config/wolf"
)

type AccountRepository struct {
	Rdb      *redis.Client
	Ctx      context.Context
	UseProxy bool
}

func (r *AccountRepository) Balance(currency string) (float64, error) {
	mutex := common.NewMutex(
		r.Rdb,
		r.Ctx,
		"locks:wolf:api",
	)
	if mutex.Lock(2 * time.Second) {
		return 0, errors.New("wolf api locked")
	}
	defer mutex.Unlock()

	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	if r.UseProxy {
		session := &common.ProxySession{
			Proxy: "socks5://127.0.0.1:1080?timeout=2s",
		}
		tr.DialContext = session.DialContext
	} else {
		session := &net.Dialer{}
		tr.DialContext = session.DialContext
	}
	httpClient := &http.Client{
		Transport: tr,
		Timeout:   2 * time.Second,
	}

	url := "https://wolf.bet/api/v1/user/balances"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.API_TOKEN))

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, errors.New(fmt.Sprintf("request error: status[%s] code[%d]", resp.Status, resp.StatusCode))
	}

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	if _, ok := data["balances"]; !ok {
		return 0, errors.New("balances not exists")
	}

	balances := make(map[string]string)

	for _, item := range data["balances"].([]interface{}) {
		balance := item.(map[string]interface{})
		if _, ok := balance["currency"]; !ok {
			return 0, errors.New("currency not exists")
		}
		if _, ok := balance["amount"]; !ok {
			return 0, errors.New("amount not exists")
		}
		balances[balance["currency"].(string)] = balance["amount"].(string)
	}

	r.Rdb.HMSet(
		r.Ctx,
		"wolf:balances",
		balances,
	)

	if _, ok := balances[currency]; !ok {
		return 0, errors.New("currency not exists")
	}

	balance, _ := strconv.ParseFloat(balances[currency], 64)

	return balance, nil
}
