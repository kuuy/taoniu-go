package commands

import (
  "fmt"
  "log"
  "net/http"
  "os"
  "taoniu.local/groceries/api"

  "github.com/go-chi/chi/v5"
  "github.com/urfave/cli/v2"

  v1 "taoniu.local/groceries/api/v1"
)

type ApiHandler struct{}

func NewApiCommand() *cli.Command {
  var h ApiHandler
  return &cli.Command{
    Name:  "api",
    Usage: "",
    Before: func(c *cli.Context) error {
      h = ApiHandler{}
      return nil
    },
    Action: func(c *cli.Context) error {
      if err := h.run(); err != nil {
        return cli.Exit(err.Error(), 1)
      }
      return nil
    },
  }
}

func (h *ApiHandler) run() error {
  log.Println("api running...")

  r := chi.NewRouter()
  r.Route("/v1", func(r chi.Router) {
    r.Use(api.Authenticator)
    r.Mount("/stores", v1.NewStoresRouter())
    r.Mount("/products", v1.NewProductsRouter())
    r.Mount("/barcodes", v1.NewBarcodesRouter())
  })

  err := http.ListenAndServe(
    fmt.Sprintf("127.0.0.1:%v", os.Getenv("GROCERIES_API_PORT")),
    r,
  )
  if err != nil {
    return err
  }

  return nil
}
