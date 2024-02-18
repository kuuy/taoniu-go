package proxies

import (
  "context"
  "github.com/go-redis/redis/v8"
  "github.com/urfave/cli/v2"
  "log"
  "strconv"
  pool "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/proxies"
  "time"
)

type TorHandler struct {
  Rdb        *redis.Client
  Ctx        context.Context
  Repository *repositories.TorRepository
}

func NewTorCommand() *cli.Command {
  var h TorHandler
  return &cli.Command{
    Name:  "tor",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = TorHandler{
        Rdb: pool.NewRedis(1),
        Ctx: context.Background(),
      }
      h.Repository = &repositories.TorRepository{
        Rdb: h.Rdb,
        Ctx: h.Ctx,
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "flush",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.flush(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "checker",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.checker(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "failed",
        Usage: "",
        Action: func(c *cli.Context) error {
          port, _ := strconv.Atoi(c.Args().Get(0))
          if port < 1 || port > 65535 {
            log.Fatal("port not in 1~65535")
            return nil
          }
          if err := h.failed(port); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "add",
        Usage: "",
        Action: func(c *cli.Context) error {
          port, _ := strconv.Atoi(c.Args().Get(0))
          if port < 1 || port > 65535 {
            log.Fatal("port not in 0~65535")
            return nil
          }
          if err := h.add(port); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "online",
        Usage: "",
        Action: func(c *cli.Context) error {
          port, _ := strconv.Atoi(c.Args().Get(0))
          if port < 1 || port > 65535 {
            log.Fatal("port not in 0~65535")
            return nil
          }
          if err := h.online(port); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "offline",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.offline(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "start",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.start(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "stop",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.stop(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "changeip",
        Usage: "",
        Action: func(c *cli.Context) error {
          port, _ := strconv.Atoi(c.Args().Get(0))
          if port <= 0 || port > 65535 {
            log.Fatal("port is not valid")
            return nil
          }
          if err := h.changeIp(port); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *TorHandler) flush() error {
  log.Println("tor flush processing...")
  items, _ := h.Rdb.SMembers(h.Ctx, "proxies:tor:offline").Result()
  timestamp := time.Now().Unix()
  for _, item := range items {
    port, _ := strconv.Atoi(item)
    err := h.Repository.ChangeIp(port)
    if err != nil {
      continue
    }
    h.Rdb.SRem(h.Ctx, "proxies:tor:offline", port)
    score, _ := h.Rdb.ZScore(h.Ctx, "proxies:tor:checker", strconv.Itoa(port)).Result()
    if score > 0 {
      continue
    }
    h.Rdb.ZAdd(
      h.Ctx,
      "proxies:tor:checker",
      &redis.Z{
        Score:  float64(timestamp),
        Member: port,
      },
    )
  }

  return nil
}

func (h *TorHandler) checker() error {
  log.Println("tor checker processing...")
  timestamp := time.Now().Unix() - 30
  items, _ := h.Rdb.ZRangeByScore(
    h.Ctx,
    "proxies:tor:checker",
    &redis.ZRangeBy{
      Min: "-inf",
      Max: strconv.FormatInt(timestamp, 10),
    },
  ).Result()
  log.Println("checker", items)
  for _, item := range items {
    port, _ := strconv.Atoi(item)
    err := h.Repository.Checker(port)
    if err != nil {
      log.Println("checker error", port, err)
    }
  }

  return nil
}

func (h *TorHandler) failed(port int) error {
  log.Println("tor failed processing...")
  return h.Repository.Failed(port)
}

func (h *TorHandler) add(port int) error {
  log.Println("tor add processing...")
  return h.Repository.Add(port)
}

func (h *TorHandler) online(port int) error {
  log.Println("tor start processing...")
  return h.Repository.Online(port)
}

func (h *TorHandler) offline() error {
  items, _ := h.Rdb.ZRangeByScore(
    h.Ctx,
    "proxies:tor:failed",
    &redis.ZRangeBy{
      Min: strconv.Itoa(3),
      Max: "+inf",
    },
  ).Result()
  for _, item := range items {
    port, _ := strconv.Atoi(item)
    h.Repository.Offline(port)
    h.Rdb.ZRem(h.Ctx, "proxies:tor:failed", port)
  }

  return nil
}

func (h *TorHandler) start() error {
  log.Println("tor start processing...")
  return h.Repository.Start()
}

func (h *TorHandler) stop() error {
  log.Println("tor stop processing...")
  return h.Repository.Stop()
}

func (h *TorHandler) changeIp(port int) error {
  log.Println("tor change ip processing...")
  return h.Repository.ChangeIp(port)
}
