package swap

import (
  "context"
  "errors"
  "fmt"
  "github.com/gagliardetto/solana-go"
  "github.com/gagliardetto/solana-go/rpc"
  "github.com/rs/xid"
  "gorm.io/gorm"
  "log"
  models "taoniu.local/cryptos/models/raydium/swap"
  "time"

  "taoniu.local/cryptos/common"
)

type TransactionsRepository struct {
  Db              *gorm.DB
  Ctx             context.Context
  MintsRepository *MintsRepository
}

func (r *TransactionsRepository) Flush(signature string) (err error) {
  walletAddr := common.GetEnvString("SOLANA_WALLET")
  if walletAddr == "" {
    return errors.New("SOLANA_WALLET not configured")
  }

  client := rpc.New(common.GetEnvString("SOLANA_RPC_URL"))

  txSig, err := solana.SignatureFromBase58(signature)
  if err != nil {
    return
  }

  txRes, err := client.GetTransaction(
    r.Ctx,
    txSig,
    &rpc.GetTransactionOpts{
      MaxSupportedTransactionVersion: &rpc.MaxSupportedTransactionVersion0,
    },
  )
  if err != nil {
    fmt.Printf("failed to get transaction %s: %v\n", signature, err)
    return
  }

  if txRes == nil {
    err = fmt.Errorf("empty transaction %s", signature)
    return
  }

  if txRes.Meta != nil && txRes.Meta.Err != nil {
    err = fmt.Errorf("transaction failed: %+v", txRes.Meta.Err)
    return
  }

  //tx, err := txRes.Transaction.GetTransaction()
  //if err != nil {
  //  return fmt.Errorf("failed to get transaction decode: %w", err)
  //}

  //txRes, err := solana.TransactionFromDecoder(bin.NewBinDecoder(out.Transaction.GetBinary()))

  // Parse swap data (simplified for now)
  err = r.parseAndSave(signature, txRes)

  return nil
}

func (r *TransactionsRepository) parseAndSave(signature string, res *rpc.GetTransactionResult) (err error) {
  // Implementation note: Full parsing of Raydium instructions is highly complex
  // and depends on the specific Raydium program version.
  // For this proposal, we implement a simplified heuristic looking for token balance changes.

  if res.BlockTime == nil {
    err = fmt.Errorf("empty block time")
    return
  }
  timestamp := time.Unix(int64(*res.BlockTime), 0)
  log.Println("timestamp", timestamp)

  // Heuristic: identify balance changes for the user's wallet
  // This captures the net effect of a swap without needing to perfectly decode instructions
  // which may change across versions.

  preBalances := res.Meta.PreTokenBalances
  postBalances := res.Meta.PostTokenBalances

  walletAddr := common.GetEnvString("SOLANA_WALLET")

  var mintIn, mintOut string
  var amountIn, amountOut float64

  changeMap := make(map[string]float64)

  for _, pre := range preBalances {
    log.Println("pre address", pre.Owner.String())
    if pre.Owner.String() == walletAddr && pre.UiTokenAmount.UiAmount != nil {
      changeMap[pre.Mint.String()] -= *pre.UiTokenAmount.UiAmount
    }
    if pre.Owner.String() == walletAddr {
      log.Println("pre amount", pre.Owner, pre.UiTokenAmount.UiAmountString)
    } else {
      log.Println("pre other amount", pre.Owner, pre.UiTokenAmount.UiAmountString)
    }
    if pre.Mint.String() == "So11111111111111111111111111111111111111112" {
      log.Println("mint", pre.UiTokenAmount.UiAmountString)
    }
  }
  for _, post := range postBalances {
    if post.Owner.String() == walletAddr && post.UiTokenAmount.UiAmount != nil {
      changeMap[post.Mint.String()] += *post.UiTokenAmount.UiAmount
    }
    if post.Owner.String() == walletAddr {
      log.Println("post amount", post.Owner, post.UiTokenAmount.UiAmountString)
    } else {
      log.Println("post other amount", post.Owner, post.UiTokenAmount.UiAmountString)
    }
    if post.Mint.String() == "So11111111111111111111111111111111111111112" {
      log.Println("mint", post.UiTokenAmount.UiAmountString)
    }
  }

  log.Println("map", changeMap)

  for mint, change := range changeMap {
    log.Println("change", mint, change)
    if change < 0 {
      mintIn = mint
      amountIn = -change
    } else if change > 0 {
      mintOut = mint
      amountOut = change
    }
  }

  if mintIn != "" || mintOut != "" {
    tx := &models.Transaction{
      ID:        xid.New().String(),
      Signature: signature,
      MintIn:    mintIn,
      AmountIn:  amountIn,
      MintOut:   mintOut,
      AmountOut: amountOut,
      Timestamp: timestamp,
      Status:    1,
    }
    // Pool address heuristic: one of the non-wallet accounts in the transaction
    // In a real Raydium swap, we'd extract it from the instruction data.
    // For now, using a placeholder or first non-wallet account.
    tx.PoolAddress = "unknown"

    r.Db.Create(tx)
    fmt.Printf("saved swap: %s -> %s (%f -> %f)\n", mintIn, mintOut, amountIn, amountOut)
  }

  return
}
