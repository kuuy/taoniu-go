package commands

import (
  "fmt"
  "github.com/go-chi/chi/v5"
  "github.com/urfave/cli/v2"
  "log"
  "net/http"
  "os"
  v1 "taoniu.local/security/api/v1"
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

func (h *ApiHandler) run() (err error) {
  log.Println("api running...")

  r := chi.NewRouter()
  r.Route("/v1", func(r chi.Router) {
    r.Mount("/gfw", v1.NewGfwRouter())
  })

  err = http.ListenAndServe(
    fmt.Sprintf("127.0.0.1:%v", os.Getenv("SECURITY_API_PORT")),
    r,
  )
  if err != nil {
    return
  }

  return nil
}
