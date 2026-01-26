package swap

import (
  "context"
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

      _, err := exchange.Buy(context.Background(), c.String("pool"), c.Uint64("amount"))
      if err != nil {
        return err
      }
      log.Println("Swap executed successfully")
      return nil
    },
  }
}
