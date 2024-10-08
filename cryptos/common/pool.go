package common

import (
  "context"
  "errors"
  "fmt"
  "strconv"
  "strings"
  "sync"
  "time"

  "database/sql"

  "github.com/eclipse/paho.golang/autopaho"
  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "github.com/nats-io/nats.go"
  "github.com/rs/xid"
  socketio "github.com/vchitai/go-socket.io/v4"
  "github.com/vchitai/go-socket.io/v4/engineio"

  "gorm.io/driver/postgres"
  "gorm.io/gorm"
)

var (
  dbPools = map[string]*sql.DB{}
)

type ApiContext struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
  Mux sync.Mutex
}

type NatsContext struct {
  Db   *gorm.DB
  Rdb  *redis.Client
  Ctx  context.Context
  Conn *nats.Conn
}

type MqttContext struct {
  Db   *gorm.DB
  Rdb  *redis.Client
  Ctx  context.Context
  Conn *autopaho.ConnectionManager
  Nats *nats.Conn
}

type AnsqServerContext struct {
  Db   *gorm.DB
  Rdb  *redis.Client
  Ctx  context.Context
  Mux  *asynq.ServeMux
  Nats *nats.Conn
}

type AnsqClientContext struct {
  Db   *gorm.DB
  Rdb  *redis.Client
  Ctx  context.Context
  Conn *asynq.Client
  Nats *nats.Conn
}

type SocketContext struct {
  Socket *socketio.Server
  Conn   socketio.Conn
  Nats   *nats.Conn
}

type Mutex struct {
  rdb   *redis.Client
  ctx   context.Context
  key   string
  value string
}

func NewRedis(i int) *redis.Client {
  return redis.NewClient(&redis.Options{
    Addr:     GetEnvString(fmt.Sprintf("REDIS_%02d_HOST", i)),
    Password: GetEnvString(fmt.Sprintf("REDIS_%02d_PASSWORD", i)),
    DB:       GetEnvInt(fmt.Sprintf("REDIS_%02d_DB", i)),
  })
}

func NewDBPool(i int) *sql.DB {
  key := fmt.Sprintf("DB_%02d_DSN", i)
  pool, ok := dbPools[key]
  if !ok || pool == nil {
    dsn := GetEnvString(key)
    var err error
    pool, err = sql.Open("pgx", dsn)
    if err != nil {
      panic(err)
    }
    pool.SetMaxIdleConns(50)
    pool.SetMaxOpenConns(100)
    pool.SetConnMaxLifetime(5 * time.Minute)
    dbPools[key] = pool
  }
  return pool
}

func NewDB(i int) *gorm.DB {
  db, err := gorm.Open(postgres.New(postgres.Config{
    Conn: NewDBPool(i),
  }), &gorm.Config{})
  if errors.Is(err, context.DeadlineExceeded) {
    return NewDB(i)
  }
  if err != nil {
    panic(err)
  }
  return db
}

func NewAsynqServer(topic string) *asynq.Server {
  rdb := asynq.RedisClientOpt{
    Addr: GetEnvString(fmt.Sprintf("ASYNQ_%s_REDIS_ADDR", topic)),
    DB:   GetEnvInt(fmt.Sprintf("ASYNQ_%s_REDIS_DB", topic)),
  }
  queues := make(map[string]int)
  for _, item := range GetEnvArray(fmt.Sprintf("ASYNQ_%s_QUEUE", topic)) {
    data := strings.Split(item, ",")
    weight, _ := strconv.Atoi(data[1])
    queues[data[0]] = weight
  }
  return asynq.NewServer(rdb, asynq.Config{
    Concurrency: GetEnvInt(fmt.Sprintf("ASYNQ_%s_CONCURRENCY", topic)),
    Queues:      queues,
  })
}

func NewAsynqClient(topic string) *asynq.Client {
  return asynq.NewClient(asynq.RedisClientOpt{
    Addr: GetEnvString(fmt.Sprintf("ASYNQ_%s_REDIS_ADDR", topic)),
    DB:   GetEnvInt(fmt.Sprintf("ASYNQ_%s_REDIS_DB", topic)),
  })
}

func NewSocketServer(opts *engineio.Options) *socketio.Server {
  server := socketio.NewServer(opts)
  _, err := server.Adapter(&socketio.RedisAdapterConfig{
    Addr:     GetEnvString("CRYPTOS_SOCKET_REDIS_ADDR"),
    Password: GetEnvString("CRYPTOS_SOCKET_REDIS_PASSWORD"),
    DB:       GetEnvInt("CRYPTOS_SOCKET_REDIS_DB"),
    Prefix:   "socket.io",
  })
  if err != nil {
    panic(err)
  }
  return server
}

func NewNats() *nats.Conn {
  nc, err := nats.Connect("127.0.0.1", nats.Token(GetEnvString("NATS_TOKEN")))
  if err != nil {
    panic(err)
  }
  return nc
}

func NewMutex(
  rdb *redis.Client,
  ctx context.Context,
  key string,
) *Mutex {
  return &Mutex{
    rdb:   rdb,
    ctx:   ctx,
    key:   key,
    value: xid.New().String(),
  }
}

func (m *Mutex) Lock(ttl time.Duration) bool {
  result, err := m.rdb.SetNX(
    m.ctx,
    m.key,
    m.value,
    ttl,
  ).Result()
  if err != nil {
    return false
  }
  return result
}

func (m *Mutex) Unlock() {
  script := redis.NewScript(`
  if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("DEL", KEYS[1])
  else
    return 0
  end
  `)
  script.Run(m.ctx, m.rdb, []string{m.key}, m.value).Result()
}
