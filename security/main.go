package main

import (
  "log"
  "os"
  "path"

  "github.com/joho/godotenv"
  "github.com/urfave/cli/v2"

  "taoniu.local/security/commands"
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

  err = app.Run(os.Args)
  if err != nil {
    log.Fatalln("error", err)
  }
}
