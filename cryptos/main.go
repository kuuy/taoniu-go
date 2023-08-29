package main

import (
  "log"
  "os"
  "path"

  "github.com/joho/godotenv"
  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/commands"
)

func main() {
  home, err := os.UserHomeDir()
  if err != nil {
    panic(err)
  }
  err = godotenv.Load(path.Join(home, "taoniu-go", ".env"))
  if err != nil {
    log.Fatal(err)
  }

  app := &cli.App{
    Name:  "cryptos commands",
    Usage: "",
    Action: func(c *cli.Context) error {
      if c.Command.Action == nil {
        cli.ShowAppHelp(c)
      } else {
        log.Fatalln("error", c.Err)
      }
      return nil
    },
    Commands: []*cli.Command{
      commands.NewApiCommand(),
      commands.NewGrpcCommand(),
      commands.NewSocketCommand(),
      commands.NewBinanceCommand(),
      commands.NewDydxCommand(),
      commands.NewCronCommand(),
      commands.NewDbCommand(),
      commands.NewQueueCommand(),
      commands.NewProxiesCommand(),
      commands.NewExchangesCommand(),
      commands.NewSectorsCommand(),
      commands.NewTradingviewCommand(),
      commands.NewParticiplesCommand(),
    },
    Version: "0.0.0",
  }

  err = app.Run(os.Args)
  if err != nil {
    log.Fatalln("error", err)
  }
}
