package v1

import (
  "net/http"
  "strconv"

  "github.com/go-chi/chi/v5"
  "gorm.io/gorm"

  "taoniu.local/groceries/api"
  "taoniu.local/groceries/common"
  "taoniu.local/groceries/repositories"
)

type StoresHandler struct {
  Db         *gorm.DB
  Response   *api.ResponseHandler
  Repository *repositories.StoresRepository
}

type StoreInfo struct {
  ID   string `json:"id"`
  Name string `json:"name"`
  Logo string `json:"logo"`
}

type StoreDetail struct {
  ID   string `json:"id"`
  Name string `json:"name"`
  Logo string `json:"logo"`
}

func NewStoresRouter() http.Handler {
  h := StoresHandler{
    Db: common.NewDB(),
  }
  h.Repository = &repositories.StoresRepository{
    Db: h.Db,
  }

  r := chi.NewRouter()
  r.Use(api.Authenticator)
  r.Get("/", h.Listings)
  r.Get("/{id:[a-z0-9]{20}}", h.Get)
  r.Post("/", h.Create)
  r.Put("/{id:[a-z0-9]{20}}", h.Update)

  return r
}

func (h *StoresHandler) Listings(w http.ResponseWriter, r *http.Request) {
  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  var current int
  if !r.URL.Query().Has("current") {
    current = 1
  }
  current, _ = strconv.Atoi(r.URL.Query().Get("current"))
  if current < 1 {
    h.Response.Error(http.StatusForbidden, 1004, "current not valid")
    return
  }

  var pageSize int
  if !r.URL.Query().Has("page_size") {
    pageSize = 50
  } else {
    pageSize, _ = strconv.Atoi(r.URL.Query().Get("page_size"))
  }
  if pageSize < 1 || pageSize > 100 {
    h.Response.Error(http.StatusForbidden, 1004, "page size not valid")
    return
  }

  conditions := make(map[string]interface{})

  total := h.Repository.Count(conditions)
  stores, err := h.Repository.Listings(conditions, current, pageSize)
  if err != nil {
    return
  }

  data := make([]*StoreInfo, len(stores))
  for i, store := range stores {
    data[i] = &StoreInfo{
      ID:   store.ID,
      Name: store.Name,
      Logo: store.Logo,
    }
  }

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)

  h.Response.Pagenate(data, total, current, pageSize)
}

func (h *StoresHandler) Get(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  id := chi.URLParam(r, "id")
  store, err := h.Repository.Find(id)
  if err != nil {
    http.Error(w, http.StatusText(404), http.StatusNotFound)
    return
  }

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)

  result := &StoreDetail{
    ID:   store.ID,
    Name: store.Name,
    Logo: store.Logo,
  }

  h.Response.Json(result)
}

func (h *StoresHandler) Create(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  r.ParseForm()

  uid := r.Header.Get("uid")
  name := r.FormValue("name")
  logo := r.FormValue("logo")

  id := h.Repository.Create(
    uid,
    name,
    logo,
  )

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)

  h.Response.Json(map[string]interface{}{
    "id": id,
  })
}

func (h *StoresHandler) Update(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  r.ParseForm()

  id := chi.URLParam(r, "id")
  _, err := h.Repository.Find(id)
  if err != nil {
    http.Error(w, http.StatusText(404), http.StatusNotFound)
    return
  }

  name := r.FormValue("name")
  logo := r.FormValue("logo")

  h.Repository.Update(
    id,
    name,
    logo,
  )

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)

  h.Response.Json(nil)
}
