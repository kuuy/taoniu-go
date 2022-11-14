package dice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"net"
	"net/http"
	"strconv"
	"taoniu.local/gamblings/common"
	"time"

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
	Auto       bool   `json:"auto"`
	Multiplier string `json:"multiplier"`
	Amount     string `json:"amount"`
}

func (r *BetRepository) Place(request *BetRequest) (string, float64, bool, error) {
	mutex := common.NewMutex(
		r.Rdb,
		r.Ctx,
		"locks:wolf:api",
	)
	if mutex.Lock(2 * time.Second) {
		return "", 0, false, errors.New("wolf api locked")
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

	body, _ := json.Marshal(request)

	url := "https://wolf.bet/api/v2/limbo/manual/play"
	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("X-Client-Type", "Web-Application")
	req.Header.Set("X-Hash-Api", config.LOGIN_HASH)
	req.Header.Set("Referer", "https://wolf.bet/casino/hilo")
	req.Header.Set("Cookie", config.LOGIN_COOKIE)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.LOGIN_TOKEN))
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", 0, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, false, errors.New(fmt.Sprintf("request error: status[%s] code[%d]", resp.Status, resp.StatusCode))
	}

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	if _, ok := data["bet"]; !ok {
		log.Println("data", data)
		return "", 0, false, errors.New("bet not exists")
	}

	bet := data["bet"].(map[string]interface{})
	if _, ok := bet["hash"]; !ok {
		return "", 0, false, errors.New("hash not exists")
	}
	if _, ok := bet["state"]; !ok {
		return "", 0, false, errors.New("state not exists")
	}
	if _, ok := bet["result_value"]; !ok {
		return "", 0, false, errors.New("result value not exists")
	}

	hash := bet["hash"].(string)
	state := false
	if bet["state"].(string) == "win" {
		state = true
	}

	result, _ := strconv.ParseFloat(bet["result_value"].(string), 64)

	return hash, result, state, nil
}
