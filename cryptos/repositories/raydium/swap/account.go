package swap

import (
  "context"
  "errors"
  "fmt"
  "math"

  bin "github.com/gagliardetto/binary"
  "github.com/gagliardetto/solana-go"
  "github.com/gagliardetto/solana-go/programs/token"
  "github.com/gagliardetto/solana-go/rpc"
  "github.com/rs/xid"
  "gorm.io/gorm"

  "taoniu.local/cryptos/common"
  models "taoniu.local/cryptos/models/raydium/swap"
)

type AccountRepository struct {
  Db              *gorm.DB
  Ctx             context.Context
  MintsRepository *MintsRepository
}

func (r *AccountRepository) Flush() (err error) {
  walletAddr := common.GetEnvString("SOLANA_WALLET")
  if walletAddr == "" {
    return errors.New("SOLANA_WALLET not configured")
  }

  pubKey, err := solana.PublicKeyFromBase58(walletAddr)
  if err != nil {
    return fmt.Errorf("invalid wallet address: %v", err)
  }

  client := rpc.New(common.GetEnvString("SOLANA_RPC_URL"))

  balances := make(map[string]float64)
  addresses := make(map[string]string)

  // 1. Flush Native SOL
  solBalance, err := client.GetBalance(r.Ctx, pubKey, rpc.CommitmentFinalized)
  if err != nil {
    return fmt.Errorf("failed to get SOL balance: %v", err)
  }
  balances["SOL"] = float64(solBalance.Value) / math.Pow10(9)
  addresses["SOL"] = walletAddr

  // 2. Flush SPL Tokens
  tokenAccounts, err := client.GetTokenAccountsByOwner(
    r.Ctx,
    pubKey,
    &rpc.GetTokenAccountsConfig{
      ProgramId: &solana.TokenProgramID,
    },
    &rpc.GetTokenAccountsOpts{
      Commitment: rpc.CommitmentFinalized,
    },
  )
  if err != nil {
    return fmt.Errorf("failed to get token accounts: %v", err)
  }

  for _, rawAccount := range tokenAccounts.Value {
    accountData := rawAccount.Account.Data.GetBinary()
    var tokenAccount token.Account
    if err = tokenAccount.UnmarshalWithDecoder(bin.NewBinDecoder(accountData)); err != nil {
      continue
    }

    mintAddr := tokenAccount.Mint.String()
    var mint *models.Mint
    mint, err = r.MintsRepository.GetByAddress(mintAddr)
    if err != nil {
      if errors.Is(err, gorm.ErrRecordNotFound) {
        continue
      }
      return
    }

    balance := float64(tokenAccount.Amount) / math.Pow10(mint.Decimals)
    balances[mint.Symbol] += balance
    if _, ok := addresses[mint.Symbol]; !ok {
      addresses[mint.Symbol] = rawAccount.Pubkey.String()
    }
  }

  // 3. Sync with database
  var existingSymbols []string
  r.Db.Model(&models.Account{}).Pluck("symbol", &existingSymbols)

  for symbol, balance := range balances {
    if err = r.saveAccount(symbol, addresses[symbol], balance); err != nil {
      return
    }
  }

  // Deactivate symbols no longer in RPC response
  var missingSymbols []string
  for _, symbol := range existingSymbols {
    if _, ok := balances[symbol]; !ok {
      missingSymbols = append(missingSymbols, symbol)
    }
  }

  if len(missingSymbols) > 0 {
    err = r.Db.Model(&models.Account{}).
      Where("symbol IN ?", missingSymbols).
      Updates(map[string]interface{}{
        "balance": 0,
        "status":  4,
      }).Error
  }

  return
}

func (r *AccountRepository) saveAccount(symbol string, address string, balance float64) error {
  var entity models.Account
  err := r.Db.Where("symbol = ?", symbol).Take(&entity).Error
  if errors.Is(err, gorm.ErrRecordNotFound) {
    entity = models.Account{
      ID:      xid.New().String(),
      Symbol:  symbol,
      Address: address,
      Balance: balance,
      Status:  1,
    }
    return r.Db.Create(&entity).Error
  } else if err != nil {
    return err
  }

  return r.Db.Model(&entity).Updates(map[string]interface{}{
    "address": address,
    "balance": balance,
    "status":  1,
  }).Error
}
