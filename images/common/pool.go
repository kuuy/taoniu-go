package common

import (
  "context"
  "time"

  "github.com/go-redis/redis/v8"
  "github.com/rs/xid"

  "database/sql"
  "gorm.io/driver/postgres"
  "gorm.io/gorm"
)

var (
  rdb    *redis.Client
  dbPool *sql.DB
  db     *gorm.DB
)

type Mutex struct {
  rdb   *redis.Client
  ctx   context.Context
  key   string
  value string
}

func NewRedis() *redis.Client {
  rdb = redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Password: "",
    DB:       8,
  })

  return rdb
}

func NewDBPool() *sql.DB {
  if dbPool == nil {
    dsn := "postgres://taoniu:64EQJMn1O9JrZ2G4@localhost/taoniu"
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
  if err != nil {
    panic(err)
  }

  return db
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
