package transactions

import (
  "context"
  "errors"
  "fmt"
  "log"

  "github.com/gagliardetto/solana-go"
  "github.com/gagliardetto/solana-go/rpc"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
)

type SignaturesRepository struct {
  Db  *gorm.DB
  Ctx context.Context
}

func (r *SignaturesRepository) Flush(limit int) (err error) {
  walletAddr := common.GetEnvString("SOLANA_WALLET")
  if walletAddr == "" {
    return errors.New("SOLANA_WALLET not configured")
  }

  pubKey, err := solana.PublicKeyFromBase58(walletAddr)
  if err != nil {
    return fmt.Errorf("invalid wallet address: %v", err)
  }

  client := rpc.New(common.GetEnvString("SOLANA_RPC_URL"))

  signatures, err := client.GetSignaturesForAddressWithOpts(
    r.Ctx,
    pubKey,
    &rpc.GetSignaturesForAddressOpts{
      Limit: &limit,
    },
  )
  if err != nil {
    return fmt.Errorf("failed to get signatures: %v", err)
  }

  for _, sigInfo := range signatures {
    log.Println("signature info", sigInfo)
  }

  return
}
