package swap

import (
  "context"

  "github.com/rs/xid"
  "gorm.io/gorm"

  models "taoniu.local/cryptos/models/raydium/swap"
)

type PositionsRepository struct {
  Db              *gorm.DB
  Ctx             context.Context
  MintsRepository *MintsRepository
}

func (r *PositionsRepository) Flush() error {
  var transactions []models.Transaction
  err := r.Db.Order("timestamp asc").Find(&transactions).Error
  if err != nil {
    return err
  }

  quoteMints := map[string]bool{
    "So11111111111111111111111111111111111111112":  true, // WSOL
    "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v": true, // USDC
    "Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB": true, // USDT
  }

  positions := make(map[string]*models.Position)

  for _, tx := range transactions {
    var tokenMint string
    var isBuy bool
    var quoteAmount, tokenAmount float64

    if quoteMints[tx.MintIn] && !quoteMints[tx.MintOut] {
      // Buy: Quote -> Token
      tokenMint = tx.MintOut
      isBuy = true
      quoteAmount = tx.AmountIn
      tokenAmount = tx.AmountOut
    } else if !quoteMints[tx.MintIn] && quoteMints[tx.MintOut] {
      // Sell: Token -> Quote
      tokenMint = tx.MintIn
      isBuy = false
      quoteAmount = tx.AmountOut
      tokenAmount = tx.AmountIn
    } else {
      // Skip transactions between two tokens or two quote currencies for now
      continue
    }

    if tokenMint == "" {
      continue
    }

    mint, err := r.MintsRepository.GetByAddress(tokenMint)
    if err != nil {
      continue
    }

    symbol := mint.Symbol
    pos, ok := positions[symbol]
    if !ok {
      // Try to load existing open position from DB
      var existing models.Position
      err = r.Db.Where("symbol = ? AND status = 1", symbol).Take(&existing).Error
      if err == nil {
        pos = &existing
      } else {
        pos = &models.Position{
          ID:     xid.New().String(),
          Symbol: symbol,
          Status: 1,
        }
      }
      positions[symbol] = pos
    }

    if isBuy {
      pos.EntryQuantity += tokenAmount
      pos.EntryAmount += quoteAmount
      if pos.EntryQuantity > 0 {
        pos.EntryPrice = pos.EntryAmount / pos.EntryQuantity
      }
      pos.Timestamp = tx.Timestamp.UnixMilli()
    } else {
      if pos.EntryQuantity > 0 {
        // Pro-rata reduction of entry amount
        ratio := tokenAmount / pos.EntryQuantity
        if ratio > 1 {
          ratio = 1
        }
        pos.EntryAmount -= pos.EntryAmount * ratio
        pos.EntryQuantity -= tokenAmount
      }
      if pos.EntryQuantity <= 0.000001 { // Handle precision issues
        pos.EntryQuantity = 0
        pos.EntryAmount = 0
        pos.Status = 2 // Closed
      }
      pos.Timestamp = tx.Timestamp.UnixMilli()
    }
  }

  for _, pos := range positions {
    if pos.ID == "" {
      continue
    }
    var count int64
    r.Db.Model(&models.Position{}).Where("id = ?", pos.ID).Count(&count)
    if count == 0 {
      r.Db.Create(pos)
    } else {
      r.Db.Save(pos)
    }
  }

  return nil
}

func (r *PositionsRepository) GetOpenPositions() (positions []models.Position, err error) {
  err = r.Db.Where("status = 1").Find(&positions).Error
  return
}
