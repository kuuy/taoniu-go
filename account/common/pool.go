package common

import (
  "context"
  "errors"
  "strconv"
  "strings"
  "time"

  "database/sql"

  "github.com/go-redis/redis/v8"
  "github.com/hibiken/asynq"
  "github.com/rs/xid"

  "gorm.io/driver/postgres"
  "gorm.io/gorm"
)

var (
  dbPool *sql.DB
)

type Mutex struct {
  rdb   *redis.Client
  ctx   context.Context
  key   string
  value string
}

func NewRedis() *redis.Client {
  return redis.NewClient(&redis.Options{
    Addr:     GetEnvString("REDIS_HOST"),
    Password: GetEnvString("REDIS_PASSWORD"),
    DB:       GetEnvInt("REDIS_DB"),
  })
}

func NewDBPool() *sql.DB {
  if dbPool == nil {
    dsn := GetEnvString("DB_DSN")
    pool, err := sql.Open("pgx", dsn)
    if err != nil {
      panic(err)
    }
    pool.SetMaxIdleConns(50)
    pool.SetMaxOpenConns(100)
    pool.SetConnMaxLifetime(5 * time.Minute)
    dbPool = pool
  }
  return dbPool
}

func NewDB() *gorm.DB {
  db, err := gorm.Open(postgres.New(postgres.Config{
    Conn: NewDBPool(),
  }), &gorm.Config{})
  if errors.Is(err, context.DeadlineExceeded) {
    return NewDB()
  }
  if err != nil {
    panic(err)
  }
  return db
}

func NewAsynqServer() *asynq.Server {
  rdb := asynq.RedisClientOpt{
    Addr: GetEnvString("ASYNQ_REDIS_ADDR"),
    DB:   GetEnvInt("ASYNQ_REDIS_DB"),
  }
  queues := make(map[string]int)
  for _, item := range GetEnvArray("ASYNQ_QUEUE") {
    data := strings.Split(item, ",")
    weight, _ := strconv.Atoi(data[1])
    queues[data[0]] = weight
  }
  return asynq.NewServer(rdb, asynq.Config{
    Concurrency: GetEnvInt("ASYNQ_CONCURRENCY"),
    Queues:      queues,
  })
}

func NewAsynqClient() *asynq.Client {
  return asynq.NewClient(asynq.RedisClientOpt{
    Addr: GetEnvString("ASYNQ_REDIS_ADDR"),
    DB:   GetEnvInt("ASYNQ_REDIS_DB"),
  })
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
  if err != redis.Nil {
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
