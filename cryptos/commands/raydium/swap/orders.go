package swap

import (
  "context"
  "fmt"
  "log"

  "github.com/gagliardetto/solana-go/rpc"
  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/common"
  "taoniu.local/cryptos/repositories/raydium/swap"
)

func NewOrdersCommand() *cli.Command {
  return &cli.Command{
    Name:  "orders",
    Usage: "Execute a swap order on Raydium",
    Flags: []cli.Flag{
      &cli.StringFlag{
        Name:     "pool",
        Usage:    "Pool address",
        Required: true,
      },
      &cli.Uint64Flag{
        Name:     "amount",
        Usage:    "Amount to swap",
        Required: true,
      },
      &cli.Float64Flag{
        Name:  "max-price",
        Usage: "Maximum price for exchange in (buy)",
      },
      &cli.Float64Flag{
        Name:  "min-price",
        Usage: "Minimum price for exchange out (sell)",
      },
      &cli.StringFlag{
        Name:  "side",
        Usage: "Side of the order (buy/sell)",
        Value: "buy",
      },
    },
    Action: func(c *cli.Context) error {
      rpcURL := common.GetEnvString("SOLANA_RPC_URL")
      if rpcURL == "" {
        rpcURL = "https://api.mainnet-beta.solana.com"
      }

      privateKey := common.GetEnvString("SOLANA_PRIVATE_KEY")
      if privateKey == "" {
        log.Println("SOLANA_PRIVATE_KEY not set")
        // Return early or error
        return cli.Exit("SOLANA_PRIVATE_KEY env var required", 1)
      }

      client := rpc.New(rpcURL)
      exchange := swap.NewExchange(client, privateKey)

      side := c.String("side")
      var sig string
      var err error
      if side == "buy" {
        sig, err = exchange.Buy(context.Background(), c.String("pool"), c.Uint64("amount"), c.Float64("max-price"))
      } else if side == "sell" {
        sig, err = exchange.Sell(context.Background(), c.String("pool"), c.Uint64("amount"), c.Float64("min-price"))
      } else {
        return cli.Exit(fmt.Sprintf("Invalid side: %s", side), 1)
      }

      if err != nil {
        return err
      }
      log.Printf("Swap executed successfully: %s", sig)
      return nil
    },
  }
}
