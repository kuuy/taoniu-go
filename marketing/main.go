package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"taoniu.local/marketing/commands"
)

func main() {
	app := &cli.App{
		Name:  "marketing commands",
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
		},
		Version: "0.0.0",
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln("error", err)
	}
}
