package main

import (
  "os"
  "log"
  "time"
  "sync"

  "github.com/urfave/cli/v2"
  "github.com/robfig/cron/v3"

  futures "taoniu.local/cryptos/tasks/binance/futures"
)

const OneSecond = 1*time.Second + 50*time.Millisecond

func main() {
  app := &cli.App{
    Name: "tasks queues",
    Usage: "taoniu cryptos tasks",
    Action: func(c *cli.Context) error {
      log.Fatalln("binance cli error", c.Err)
      return nil
    },
    Commands: []*cli.Command{
      {
        Name: "run",
        Usage: "run tasks",
        Action: func(c *cli.Context) error {
          if err := run(); err != nil {
            return cli.NewExitError(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name: "test",
        Usage: "test tasks",
        Action: func(c *cli.Context) error {
          if err := test(); err != nil {
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

func wait(wg *sync.WaitGroup) chan bool {
	ch := make(chan bool)
	go func() {
		wg.Wait()
		ch <- true
	}()
	return ch
}

func test() error {
  log.Println("test tasks")


  futures.StochRsi()
  futures.SubmitOrder()
  err := futures.TakeProfit()
  if err != nil {
    log.Fatalln("test failed", err)
  }

  return nil
}

func run() error {
  log.Println("run tasks")

  wg := &sync.WaitGroup{}
	wg.Add(1)

  c := cron.New()
  c.AddFunc("@every 1s", func() {
    futures.FlushKline5s()
  })
  c.AddFunc("@every 10s", func() {
    futures.StochRsi()
    futures.SubmitOrder()
    futures.TakeProfit()
  })
  c.AddFunc("@every 5s", func() {
    futures.Pivot()
  })
  c.AddFunc("@every 20s", func() {
    futures.FlushRules()
  })
  c.AddFunc("@every 30s", func() {
    futures.FlushAccount()
    futures.FlushOrders()
  })
  c.AddFunc("@every 1h", func() {
    futures.CleanKline5s()
  })
  c.Start()

	<-wait(wg)

  return nil
}

