package main

import (
  "log"
  "os"
  "path"
  "path/filepath"

  "github.com/joho/godotenv"
  "github.com/urfave/cli/v2"

  "taoniu.local/account/commands"
)

func main() {
  if err := godotenv.Load(path.Join(filepath.Dir(os.Args[0]), ".env")); err != nil {
    dir, _ := os.Getwd()
    if err = godotenv.Load(path.Join(dir, ".env")); err != nil {
      panic(err)
    }
  }

  app := &cli.App{
    Name:  "account commands",
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
      commands.NewDbCommand(),
      commands.NewUsersCommand(),
      commands.NewMqttCommand(),
    },
    Version: "0.0.0",
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatalln("error", err)
  }
}
