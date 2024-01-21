package commands

import (
  "context"
  "github.com/eclipse/paho.golang/autopaho"
  "github.com/eclipse/paho.golang/paho"
  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "gorm.io/gorm"
  "log"
  "net/url"
  "taoniu.local/cryptos/common"
  "time"
)

type MqttHandler struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func NewMqttCommand() *cli.Command {
  var h MqttHandler
  return &cli.Command{
    Name:  "mqtt",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = MqttHandler{
        Db:  common.NewDB(),
        Rdb: common.NewRedis(),
        Ctx: context.Background(),
      }
      return nil
    },
    Action: func(c *cli.Context) error {
      token := c.Args().Get(0)
      if token == "" {
        log.Fatal("token is empty")
        return nil
      }
      if err := h.run(token); err != nil {
        return cli.Exit(err.Error(), 1)
      }
      return nil
    },
  }
}

func (h *MqttHandler) run(token string) (err error) {
  log.Println("mqtt running...")

  clientID := "taoniu-go"

  broker, err := url.Parse("mqtt://127.0.0.1:1883/")
  mqttcfg := autopaho.ClientConfig{
    BrokerUrls: []*url.URL{broker},
    OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
      log.Println("mqtt connection up")
    },
    OnConnectError: func(err error) {
      log.Println("mqtt connect", "err", err)
    },
    ClientConfig: paho.ClientConfig{
      ClientID: clientID,
      OnClientError: func(err error) {
        log.Println("mqtt server requested disconnect (client error)", "err", err)
      },
      OnServerDisconnect: func(d *paho.Disconnect) {
        if d.Properties != nil {
          log.Println("mqtt server requested disconnect", "reason", d.Properties.ReasonString)
        } else {
          log.Println("mqtt server requested disconnect", "reasonCode", d.ReasonCode)
        }
      },
      Router: paho.NewSingleHandlerRouter(func(m *paho.Publish) {
        log.Println("mqtt message (unhandled)", "topic", m.Topic, "payload", m.Payload)
      }),
    },
  }
  mqttcfg.SetUsernamePassword(token, []byte("jwt"))

  cm, err := autopaho.NewConnection(h.Ctx, mqttcfg)
  if err != nil {
    log.Println("connect mqtt server failed.", err)
    return
  }
  ctx, _ := context.WithTimeout(h.Ctx, time.Duration(1)*time.Second)
  err = cm.AwaitConnection(ctx)
  if err != nil {
    log.Println("err", err)
  }
  log.Println("connect", cm)
  //go func() {
  //  for {
  //    select {
  //    case <-time.After(1 * time.Hour):
  //      log.Println("online message publish")
  //      //publishOnlineMessage(cm)
  //    case <-cm.Done():
  //      return
  //    }
  //  }
  //}()

  //conn, err := net.Dial("tcp", "127.0.0.1:1883")
  //if err != nil {
  //  return
  //}
  //c := paho.NewClient(paho.ClientConfig{
  //  Conn: conn,
  //})
  //cp := &paho.Connect{
  //  KeepAlive:  30,
  //  ClientID:   "taoniu-go",
  //  CleanStart: true,
  //  Username:   "qubing@kuuy.com",
  //  Password:   []byte("my20080810#"),
  //}
  //ca, err := c.Connect(h.Ctx, cp)
  //if err != nil {
  //  return
  //}
  //if ca.ReasonCode != 0 {
  //  log.Fatalf("Failed to connect to mqtt server : %d - %s", ca.ReasonCode, ca.Properties.ReasonString)
  //  return
  //}

  //
  //c := paho.NewClient(paho.ClientConfig{
  //  //OnPublishReceived: []func(paho.PublishReceived) (bool, error){
  //  //  func(pr paho.PublishReceived) (bool, error) {
  //  //    log.Printf("%s : %s", pr.Packet.Properties.User.Get("chatname"), string(pr.Packet.Payload))
  //  //    return true, nil
  //  //  }},
  //  Conn: conn,
  //})

  //
  //cp := &paho.Connect{
  //  KeepAlive:  30,
  //  ClientID:   "taoniu-go",
  //  CleanStart: true,
  //}
  //ca, err := c.Connect(h.Ctx, cp)
  //if err != nil {
  //  return err
  //}
  //if ca.ReasonCode != 0 {
  //  log.Fatalf("Failed to connect to mqtt server : %d - %s", ca.ReasonCode, ca.Properties.ReasonString)
  //}
  //
  //ic := make(chan os.Signal, 1)
  //signal.Notify(ic, os.Interrupt, syscall.SIGTERM)
  //go func() {
  //  <-ic
  //  fmt.Println("signal received, exiting")
  //  if c != nil {
  //    d := &paho.Disconnect{ReasonCode: 0}
  //    err := c.Disconnect(d)
  //    if err != nil {
  //      log.Fatalf("failed to send Disconnect: %s", err)
  //    }
  //  }
  //  os.Exit(0)
  //}()
  //
  //if _, err := c.Subscribe(context.Background(), &paho.Subscribe{
  //  Subscriptions: []paho.SubscribeOptions{
  //    {Topic: "chat", QoS: byte(0), NoLocal: true},
  //  },
  //}); err != nil {
  //  log.Fatalln(err)
  //}
  //
  //stdin := bufio.NewReader(os.Stdin)
  //
  //for {
  //  message, err := stdin.ReadString('\n')
  //  if err == io.EOF {
  //    os.Exit(0)
  //  }
  //
  //  props := &paho.PublishProperties{}
  //  props.User.Add("chatname", "hadi")
  //
  //  pb := &paho.Publish{
  //    Topic:      "chat",
  //    QoS:        byte(0),
  //    Payload:    []byte(message),
  //    Properties: props,
  //  }
  //
  //  if _, err = c.Publish(context.Background(), pb); err != nil {
  //    log.Println(err)
  //  }
  //}

  //s := mqtt.NewServer()
  //
  //lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%v", os.Getenv("CRYPTOS_GRPC_PORT")))
  //if err != nil {
  //  log.Fatalf("net.Listen err: %v", err)
  //}
  //
  //services.NewBinance(h.Db, h.Rdb, h.Ctx).Register(s)
  //
  //s.Serve(lis)

  return nil
}
