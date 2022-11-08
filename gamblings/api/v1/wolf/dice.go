package wolf

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"taoniu.local/gamblings/api/v1/wolf/dice"
)

func NewDiceRouter() http.Handler {
	r := chi.NewRouter()
	r.Mount("/hunts", dice.NewHuntRouter())

	return r
}
