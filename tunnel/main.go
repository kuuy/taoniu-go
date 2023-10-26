package main

import (
  "fmt"
  "log"
  "net/http"
  "os"
  "path"
  "strings"

  "github.com/joho/godotenv"

  "taoniu.local/tunnel/repositories"
)

type Handler struct {
  ContentRepository *repositories.ContentRepository
}

func (h *Handler) process(w http.ResponseWriter, r *http.Request) {
  if r.URL.Path == "" || r.URL.Path == "/" {
    os.Exit(1)
  }
  segments := strings.Split(r.URL.Path, "/")
  domain := segments[1]
  path := strings.Join(segments[2:], "/")

  log.Println("query", r.URL.RawQuery)

  h.ContentRepository = &repositories.ContentRepository{
    Method:  r.Method,
    Domain:  domain,
    Path:    path,
    Query:   r.URL.RawQuery,
    Body:    r.Body,
    Timeout: 10,
  }
  h.ContentRepository.Process(w)
}

func main() {
  home, err := os.UserHomeDir()
  if err != nil {
    panic(err)
  }
  err = godotenv.Load(path.Join(home, "taoniu-go", ".env"))
  if err != nil {
    log.Fatal(err)
  }

  h := &Handler{}

  mux := http.NewServeMux()
  mux.HandleFunc("/", h.process)

  http.ListenAndServe(
    fmt.Sprintf("127.0.0.1:%v", os.Getenv("TUNNEL_PORT")),
    mux,
  )
}
