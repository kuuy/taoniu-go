package swap

import (
  "context"
  "fmt"
  "log"

  "github.com/gagliardetto/solana-go"
  "github.com/gagliardetto/solana-go/rpc"
  "github.com/shopspring/decimal"
)

type Exchange struct {
  client     *rpc.Client
  repository *OrdersRepository
  privateKey solana.PrivateKey
}

func NewExchange(client *rpc.Client, privateKeyStr string) *Exchange {
  privateKey, err := solana.PrivateKeyFromBase58(privateKeyStr)
  if err != nil {
    log.Printf("Invalid private key: %v", err)
  }
  return &Exchange{
    client:     client,
    repository: NewOrdersRepository(client),
    privateKey: privateKey,
  }
}

func (e *Exchange) Buy(ctx context.Context, poolAddress string, amountIn uint64, maxPrice float64) (string, error) {
  // 1. Get pool info from repository
  price, err := e.repository.GetPoolPrice(ctx, poolAddress)
  if err != nil {
    return "", err
  }

  if maxPrice > 0 {
    if decimal.NewFromFloat(price).GreaterThan(decimal.NewFromFloat(maxPrice)) {
      return "", fmt.Errorf("current price %v is higher than max price %v", price, maxPrice)
    }
  }

  // 2. Construct swap instruction (Placeholder)
  // In a real implementation:
  // inst := raydium.NewSwapInstruction(...)
  // tx, err := solanago.NewTransaction(
  //    []solanago.Instruction{inst},
  //    recentBlockhash,
  //    solanago.TransactionPayer(e.privateKey.PublicKey()),
  // )

  // 3. Sign and send
  // sig, err := e.client.RPC().SendTransaction(ctx, tx)

  log.Printf("Would execute BUY on pool %s with amount %d", poolAddress, amountIn)
  return "simulated_tx_signature", nil
}

func (e *Exchange) Sell(ctx context.Context, poolAddress string, amountIn uint64, minPrice float64) (string, error) {
  // 1. Get pool info from repository
  price, err := e.repository.GetPoolPrice(ctx, poolAddress)
  if err != nil {
    return "", err
  }

  if minPrice > 0 {
    if decimal.NewFromFloat(price).LessThan(decimal.NewFromFloat(minPrice)) {
      return "", fmt.Errorf("current price %v is lower than min price %v", price, minPrice)
    }
  }

  log.Printf("Would execute SELL on pool %s with amount %d", poolAddress, amountIn)
  return "simulated_tx_signature", nil
}
