package main

import (
  "os"
	"context"
	"errors"
	"fmt"
	"time"
  "log"
  "strconv"

	"io/ioutil"
	"net/http"

  "github.com/urfave/cli/v2"
	"github.com/bitly/go-simplejson"
	"github.com/rs/xid"

  //"github.com/RichardKnop/machinery/v2/log"

	"gorm.io/gorm"

	future "taoniu.local/cryptos/models"
	pool "taoniu.local/cryptos/common"
)

func main() {
  app := &cli.App{
    Name: "binance futures rules",
    Usage: "",
    Action: func(c *cli.Context) error {
      log.Fatalln("error", c.Err)
      return nil
    },
    Commands: []*cli.Command{
      {
        Name: "flush",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := flush(); err != nil {
            return cli.NewExitError(err.Error(), 1)
          }
          return nil
        },
      },
    },
    Version: "0.0.0",
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatalln("app start fatal", err)
  }
}

func newJSON(data []byte) (j *simplejson.Json, err error) {
  j, err = simplejson.NewJson(data)
  if err != nil {
    return nil, err
  }
  return j, nil
}

func flush() error {
  ctx := context.Background()
  rdb := pool.NewRedis()
  defer rdb.Close()
  db := pool.NewDB()

  mutex := pool.NewMutex(
    rdb,
    ctx,
    "lock:binance:futures:rules:flush",
  )
  if mutex.Lock(30 * time.Second) {
    return nil
  }
  defer mutex.Unlock()

  url := "https://www.binance.com/bapi/asset/v2/public/asset-service/product/get-products?includeEtf=true"
  resp, err := http.Get(url)
  if err != nil {
    fmt.Println("http request error", err)
    panic(err)
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    fmt.Printf("http request failed")
    return nil
  }

  content, _ := ioutil.ReadAll(resp.Body)
  j, _ := newJSON(content)
  data := j.Get("data").MustArray()
  timestamp := time.Now().Unix()

  for _, row := range data {
    item := row.(map[string]interface{})

    if item["q"].(string) != "USDT" {
      continue
    }

    symbol := item["s"].(string)
    price,_ := strconv.ParseFloat(item["c"].(string), 64)
    open,_ := strconv.ParseFloat(item["o"].(string), 64)
    high,_ := strconv.ParseFloat(item["h"].(string), 64)
    low,_ := strconv.ParseFloat(item["l"].(string), 64)
    volume,_ := strconv.ParseFloat(item["v"].(string), 64)
    quota,_ := strconv.ParseFloat(item["qv"].(string), 64)
    quantityStep,_ := strconv.ParseFloat(item["i"].(string), 64)
    ticketStep,_ := strconv.ParseFloat(item["ts"].(string), 64)

    rdb.HMSet(
      ctx,
      fmt.Sprintf("binance:futures:realtime:%s", item["s"]),
      map[string]interface{} {
        "symbol": symbol,
        "price": price,
        "open": open,
        "high": high,
        "low": low,
        "volume": volume,
        "quota": quota,
        "quantity_step":quantityStep,
        "ticket_step":ticketStep,
        "timestamp": fmt.Sprint(timestamp), 
      },
    )

    var entity future.Future
    result := db.Where("symbol", item["s"]).First(&entity)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
      entity = future.Future{
        ID:xid.New().String(),
        Symbol:symbol,
        Price:price,
        Open:open,
        High:high,
        Low:low,
        Volume:volume,
        Quota:quota,
        TicketStep:ticketStep,
      }
      db.Create(&entity)
    } else {
      entity.Price = price
      entity.Open = open
      entity.High = high
      entity.Low = low
      entity.Volume = volume
      entity.Quota = quota
      entity.TicketStep = ticketStep
      db.Model(&future.Future{ID:entity.ID}).Updates(entity)
    }
    fmt.Println(entity)
  }

  return nil
}

