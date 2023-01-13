package main

import (
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"taoniu.local/bt/commands"
)

func main() {
	app := &cli.App{
		Name:  "bt commands",
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
			commands.NewDbCommand(),
			commands.NewDhtCommands(),
			commands.NewTrackersCommands(),
		},
		Version: "0.0.0",
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln("error", err)
	}
}
