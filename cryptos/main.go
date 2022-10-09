package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"taoniu.local/cryptos/commands"
)

func main() {
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
			commands.NewBinanceCommand(),
			commands.NewCronCommand(),
			commands.NewDbCommand(),
		},
		Version: "0.0.0",
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln("error", err)
	}
}
