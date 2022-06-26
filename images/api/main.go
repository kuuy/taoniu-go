package main

import (
	"log"
	"net/http"

	"taoniu.local/images/api/routers"
)

func main() {
  log.Println("start api service")

  r := routers.NewImageRouter()

  http.ListenAndServe(":3080", r)
}

