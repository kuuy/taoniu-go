package swap

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"taoniu.local/cryptos/common"
	config "taoniu.local/cryptos/config/raydium/swap"
)

type TickersRepository struct {
	Db  *gorm.DB
	Rdb *redis.Client
	Ctx context.Context
}

func (r *TickersRepository) Flush() error {
	tr := &http.Transport{
		DisableKeepAlives: true,
	}

	proxy := common.GetEnvString("RAYDIUM_PROXY")
	if proxy != "" {
		tr.DialContext = (&common.ProxySession{
			Proxy: fmt.Sprintf("%v?timeout=30s", proxy),
		}).DialContext
	} else {
		tr.DialContext = (&net.Dialer{}).DialContext
	}

	httpClient := &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}

	// Using the pool info list API from Raydium V3
	url := "https://api-v3.raydium.io/main/pool/list?pageSize=100&page=1"
	fmt.Printf("fetching tickers from: %s\n", url)
	req, err := http.NewRequestWithContext(r.Ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request error: status[%s] code[%d]", resp.Status, resp.StatusCode)
	}

	var response TickersListingsResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("decode error: %v", err)
	}

	if !response.Success {
		return fmt.Errorf("failed to fetch tickers: api returned success = false")
	}

	timestamp := time.Now().UnixMilli()
	pipe := r.Rdb.Pipeline()
	for _, ticker := range response.Data.Data {
		redisKey := fmt.Sprintf(config.REDIS_KEY_TICKERS, ticker.PoolId)
		pipe.HMSet(
			r.Ctx,
			redisKey,
			map[string]interface{}{
				"pool_id":      ticker.PoolId,
				"symbol":       ticker.Symbol,
				"price":        ticker.Price,
				"volume_24h":   ticker.Volume24h,
				"change_24h":   ticker.PriceChange,
				"last_updated": ticker.LastUpdateAt,
				"timestamp":    timestamp,
			},
		)
		// Also store by symbol for easier lookup if needed
		symbolKey := fmt.Sprintf(config.REDIS_KEY_TICKERS, ticker.Symbol)
		pipe.HMSet(
			r.Ctx,
			symbolKey,
			map[string]interface{}{
				"pool_id":      ticker.PoolId,
				"symbol":       ticker.Symbol,
				"price":        ticker.Price,
				"volume_24h":   ticker.Volume24h,
				"change_24h":   ticker.PriceChange,
				"last_updated": ticker.LastUpdateAt,
				"timestamp":    timestamp,
			},
		)
	}
	_, err = pipe.Exec(r.Ctx)
	if err != nil {
		return fmt.Errorf("redis pipe exec error: %v", err)
	}

	fmt.Printf("flushed %d tickers to redis\n", len(response.Data.Data))
	return nil
}
