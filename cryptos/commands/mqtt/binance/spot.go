package binance

import (
  "context"
  "log"
  "net/url"
  "sync"
  "time"

  "github.com/eclipse/paho.golang/autopaho"
  "github.com/eclipse/paho.golang/paho"
  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  services "taoniu.local/cryptos/grpc/services/account/mqtt"
  workers "taoniu.local/cryptos/mqtt/workers/binance"
  repositories "taoniu.local/cryptos/repositories/mqtt"
)

type SpotHandler struct {
  Db                   *gorm.DB
  Rdb                  *redis.Client
  Ctx                  context.Context
  PublishersRepository *repositories.PublishersRepository
}

func NewSpotCommand() *cli.Command {
  var h SpotHandler
  return &cli.Command{
    Name:  "spot",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = SpotHandler{
        Db:  common.NewDB(1),
        Rdb: common.NewRedis(1),
        Ctx: context.Background(),
      }
      h.PublishersRepository = &repositories.PublishersRepository{}
      h.PublishersRepository.Service = &services.Publishers{
        Ctx: h.Ctx,
      }
      return nil
    },
    Action: func(c *cli.Context) error {
      if err := h.Run(); err != nil {
        return cli.Exit(err.Error(), 1)
      }
      return nil
    },
  }
}

func (h *SpotHandler) Run() (err error) {
  log.Println("mqtt running...")

  wg := &sync.WaitGroup{}
  wg.Add(1)

  token, err := h.PublishersRepository.Token(common.GetEnvString("MQTT_CRYPTOS_PUBLISHER_ID"))
  if err != nil {
    return
  }

  broker, err := url.Parse(common.GetEnvString("MQTT_BROKER_URL"))
  cfg := autopaho.ClientConfig{
    BrokerUrls: []*url.URL{broker},
    KeepAlive:  20,
    OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
      log.Println("mqtt connection up")
    },
    OnConnectError: func(err error) {
      log.Println("mqtt connect", "err", err)
    },
    ClientConfig: paho.ClientConfig{
      ClientID: common.GetEnvString("MQTT_CRYPTOS_CLIENT_ID"),
      OnClientError: func(err error) {
        log.Println("mqtt server requested disconnect (client error)", "err", err)
      },
      Router: paho.NewSingleHandlerRouter(func(m *paho.Publish) {
        log.Println("mqtt message (unhandled)", "topic", m.Topic, "payload", m.Payload)
      }),
    },
  }
  cfg.SetUsernamePassword(token.AccessToken, []byte("jwt"))

  cm, err := autopaho.NewConnection(h.Ctx, cfg)
  if err != nil {
    log.Println("connect mqtt server failed.", err)
    return
  }

  ctx, _ := context.WithTimeout(h.Ctx, time.Duration(1)*time.Second)
  err = cm.AwaitConnection(ctx)
  if err != nil {
    log.Println("err", err)
  }

  nc := common.NewNats()
  defer nc.Close()

  mqttContext := &common.MqttContext{
    Db:   h.Db,
    Rdb:  h.Rdb,
    Ctx:  h.Ctx,
    Conn: cm,
    Nats: nc,
  }

  workers.NewSpot(mqttContext).Subscribe()

  //if _, err := cm.Subscribe(h.Ctx, &paho.Subscribe{
  //  Subscriptions: []paho.SubscribeOptions{
  //    {
  //      Topic: "binance/spot/tickers/#",
  //      QoS:   0,
  //    },
  //  },
  //}); err != nil {
  //  log.Fatalf("failed to subscribe: %s", err)
  //}

  <-h.wait(wg)

  return
}

func (h *SpotHandler) wait(wg *sync.WaitGroup) chan bool {
  ch := make(chan bool)
  go func() {
    wg.Wait()
    ch <- true
  }()
  return ch
}
