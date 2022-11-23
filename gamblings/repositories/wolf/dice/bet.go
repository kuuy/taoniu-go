package dice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"

	"taoniu.local/gamblings/common"
	config "taoniu.local/gamblings/config/wolf"
)

type BetRepository struct {
	Rdb      *redis.Client
	Ctx      context.Context
	UseProxy bool
}

type BetRequest struct {
	Currency   string `json:"currency"`
	Game       string `json:"game"`
	Multiplier string `json:"multiplier"`
	Amount     string `json:"amount"`
	Rule       string `json:"rule"`
	BetValue   string `json:"bet_value"`
}

func (r *BetRepository) BetValue(rule string, multiplier float64) float64 {
	if rule == "under" {
		return math.Round(9900/multiplier) / 100
	}

	if rule == "over" {
		return 99.99 - math.Round(9900/multiplier)/100
	}

	return 0
}

func (r *BetRepository) Multiplier(rule string, betValue float64) float64 {
	if rule == "under" {
		return math.Round(990000/betValue) / 10000
	}

	if rule == "over" {
		return math.Round(990000/(100-betValue-0.01)) / 10000
	}

	return 0
}

func (r *BetRepository) Start() {
	timestamp := time.Now().Unix()
	r.Rdb.ZAdd(r.Ctx, "wolf:bet", &redis.Z{
		Score:  float64(timestamp),
		Member: "dice",
	})
}

func (r *BetRepository) Stop() {
	r.Rdb.ZRem(r.Ctx, "wolf:bet", "dice")
}

func (r *BetRepository) Place(currency string, rule string, amount float64, multiplier float64) (string, float64, float64, error) {
	mutex := common.NewMutex(
		r.Rdb,
		r.Ctx,
		"locks:wolf:api",
	)
	if mutex.Lock(2 * time.Second) {
		return "", 0, 0, errors.New("wolf api locked")
	}
	defer mutex.Unlock()

	betValue := r.BetValue(rule, multiplier)
	multiplier = r.Multiplier(rule, betValue)

	request := &BetRequest{
		Currency:   currency,
		Game:       "dice",
		Multiplier: strconv.FormatFloat(multiplier, 'f', -1, 64),
		Amount:     strconv.FormatFloat(amount, 'f', -1, 64),
		Rule:       rule,
		BetValue:   strconv.FormatFloat(betValue, 'f', -1, 64),
	}

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

	body, _ := json.Marshal(request)

	url := "https://wolf.bet/api/v1/bet/place"
	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.API_TOKEN))
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, 0, errors.New(fmt.Sprintf("request error: status[%s] code[%d]", resp.Status, resp.StatusCode))
	}

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	if _, ok := data["bet"]; !ok {
		return "", 0, 0, errors.New("bet not exists")
	}

	bet := data["bet"].(map[string]interface{})
	if _, ok := bet["hash"]; !ok {
		return "", 0, 0, errors.New("hash not exists")
	}
	if _, ok := bet["result_value"]; !ok {
		return "", 0, 0, errors.New("result value not exists")
	}
	if _, ok := bet["profit"]; !ok {
		return "", 0, 0, errors.New("state not exists")
	}

	hash := bet["hash"].(string)

	result, _ := strconv.ParseFloat(bet["result_value"].(string), 64)
	profit, _ := strconv.ParseFloat(bet["profit"].(string), 64)

	return hash, result, profit, nil
}
