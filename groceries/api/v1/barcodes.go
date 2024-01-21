package v1

import (
  "fmt"
  "net/http"

  "github.com/go-chi/chi/v5"
  "gorm.io/gorm"

  "taoniu.local/groceries/api"
  "taoniu.local/groceries/common"
  "taoniu.local/groceries/repositories"
)

type BarcodesHandler struct {
  Db                 *gorm.DB
  Response           *api.ResponseHandler
  Repository         *repositories.BarcodesRepository
  ProductsRepository *repositories.ProductsRepository
}

type BarcodeInfo struct {
  ID        string `json:"id"`
  Barcode   string `json:"barcode"`
  ProductID string `json:"product_id"`
}

func NewBarcodesRouter() http.Handler {
  h := BarcodesHandler{
    Db: common.NewDB(),
  }
  h.Repository = &repositories.BarcodesRepository{
    Db: h.Db,
  }
  h.ProductsRepository = &repositories.ProductsRepository{
    Db: h.Db,
  }

  r := chi.NewRouter()
  r.Use(api.Authenticator)
  r.Get("/", h.Get)
  r.Post("/", h.Create)
  r.Put("/", h.Update)

  return r
}

func (h *BarcodesHandler) Get(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  uid := r.Header.Get("uid")
  barcode := chi.URLParam(r, "barcode")
  info, err := h.Repository.Get(uid, barcode)
  if err != nil {
    http.Error(w, http.StatusText(404), http.StatusNotFound)
    return
  }

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)

  result := &BarcodeInfo{
    ID:        info.ID,
    Barcode:   info.Barcode,
    ProductID: info.ProductID,
  }

  h.Response.Json(result)
}

func (h *BarcodesHandler) Create(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  r.ParseForm()

  productID := r.FormValue("product_id")
  barcode := r.FormValue("barcode")

  entity, err := h.ProductsRepository.Find(productID)
  if err != nil {
    h.Response.Error(http.StatusInternalServerError, 404, "product not exists")
    return
  }

  uid := r.Header.Get("uid")
  if uid != entity.Uid {
    h.Response.Error(http.StatusInternalServerError, 404, "uid not match")
    return
  }

  _, err = h.Repository.Get(uid, barcode)
  if err == nil {
    h.Response.Error(http.StatusInternalServerError, 403, "barcode exists")
    return
  }

  id := h.Repository.Create(
    uid,
    productID,
    barcode,
  )

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)

  h.Response.Json(map[string]interface{}{
    "id": id,
  })
}

func (h *BarcodesHandler) Update(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  r.ParseForm()

  id := r.FormValue("id")
  barcode := r.FormValue("barcode")

  entity, err := h.Repository.Find(id)
  if err != nil {
    h.Response.Error(http.StatusInternalServerError, 404, "barcode not exists")
    return
  }

  uid := r.Header.Get("uid")
  if uid != entity.Uid {
    h.Response.Error(http.StatusInternalServerError, 404, "uid not match")
    return
  }

  entity, err = h.Repository.Get(uid, barcode)
  if err == nil && id != entity.ID {
    h.Response.Error(
      http.StatusInternalServerError,
      403,
      fmt.Sprintf("barcode conflict with product %v", entity.ProductID),
    )
    return
  }

  h.Repository.Update(
    id,
    barcode,
  )

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
}
