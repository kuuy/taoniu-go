package savings

import (
  "context"
  "errors"
  "log"

  "github.com/urfave/cli/v2"

  "taoniu.local/cryptos/common"
  repositories "taoniu.local/cryptos/repositories/binance/savings"
)

type ProductsHandler struct {
  Repository *repositories.ProductsRepository
}

func NewProductsCommand() *cli.Command {
  var h ProductsHandler
  return &cli.Command{
    Name:  "products",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = ProductsHandler{}
      h.Repository = &repositories.ProductsRepository{
        Db:  common.NewDB(1),
        Ctx: context.Background(),
      }
      return nil
    },
    Subcommands: []*cli.Command{
      {
        Name:  "flush",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Flush(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
      {
        Name:  "purchase",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := h.Purchase(); err != nil {
            return cli.Exit(err.Error(), 1)
          }
          return nil
        },
      },
    },
  }
}

func (h *ProductsHandler) Flush() error {
  log.Println("savings products flush...")
  return h.Repository.Flush()
}

func (h *ProductsHandler) Purchase() error {
  log.Println("savings products purchase...")
  asset := "LAZIO"
  amount := 0.1
  product, err := h.Repository.Get(asset)
  if err != nil {
    return err
  }
  if product.Status != "PURCHASING" {
    return errors.New("product not available")
  }
  if product.MinPurchaseAmount > amount {
    return errors.New("amount a bit little")
  }
  purchaseId, err := h.Repository.Purchase(product.ProductId, 0.1)
  if err != nil {
    return err
  }
  log.Println("purchaseId", purchaseId)
  return nil
}
