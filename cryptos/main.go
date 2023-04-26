package main

import (
	"github.com/joho/godotenv"
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"taoniu.local/cryptos/commands"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
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
			commands.NewBinanceCommand(),
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
