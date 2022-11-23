package wolf

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"taoniu.local/gamblings/api/v1/wolf/dice"
)

func NewDiceRouter() http.Handler {
	r := chi.NewRouter()
	r.Mount("/hunt", dice.NewHuntRouter())
	r.Mount("/plans", dice.NewPlansRouter())
	r.Mount("/hells", dice.NewHellsRouter())
	r.Mount("/multiple", dice.NewMultipleRouter())

	return r
}
