package main

import (
  "log"
  "os"
  "path"
  "path/filepath"

  "github.com/joho/godotenv"
  "github.com/urfave/cli/v2"

  "taoniu.local/security/commands"
)

func main() {
  if err := godotenv.Load(path.Join(filepath.Dir(os.Args[0]), ".env")); err != nil {
    dir, _ := os.Getwd()
    if err = godotenv.Load(path.Join(dir, ".env")); err != nil {
      panic(err)
    }
  }

  app := &cli.App{
    Name:  "security commands",
    Usage: "",
    Action: func(c *cli.Context) error {
      if c.Command.Action == nil {
        err := cli.ShowAppHelp(c)
        if err != nil {
          return err
        }
      } else {
        log.Fatalln("error", c.Err())
      }
      return nil
    },
    Commands: []*cli.Command{
      commands.NewDbCommand(),
      commands.NewApiCommand(),
      commands.NewCronCommand(),
      commands.NewGfwCommand(),
      commands.NewTorCommand(),
    },
    Version: "0.0.0",
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatalln("error", err)
  }
}
