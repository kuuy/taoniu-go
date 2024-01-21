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

type ProductsHandler struct {
  Db         *gorm.DB
  Response   *api.ResponseHandler
  Repository *repositories.ProductsRepository
}

type ProductInfo struct {
  ID    string  `json:"id"`
  Title string  `json:"title"`
  Intro string  `json:"intro"`
  Price float64 `json:"price"`
  Cover string  `json:"cover"`
}

type ProductDetail struct {
  ID    string  `json:"id"`
  Title string  `json:"title"`
  Intro string  `json:"intro"`
  Price float64 `json:"price"`
  Cover string  `json:"cover"`
}

func NewProductsRouter() http.Handler {
  h := ProductsHandler{
    Db: common.NewDB(),
  }
  h.Repository = &repositories.ProductsRepository{
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

func (h *ProductsHandler) Listings(w http.ResponseWriter, r *http.Request) {
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
  entities, err := h.Repository.Listings(conditions, current, pageSize)
  if err != nil {
    return
  }

  data := make([]*ProductInfo, len(entities))
  for i, product := range entities {
    data[i] = &ProductInfo{
      ID:    product.ID,
      Title: product.Title,
      Intro: product.Intro,
      Price: product.Price,
      Cover: product.Cover,
    }
  }

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)

  h.Response.Pagenate(data, total, current, pageSize)
}

func (h *ProductsHandler) Get(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  id := chi.URLParam(r, "id")
  product, err := h.Repository.Find(id)
  if err != nil {
    http.Error(w, http.StatusText(404), http.StatusNotFound)
    return
  }

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)

  result := &ProductDetail{
    ID:    product.ID,
    Title: product.Title,
    Intro: product.Intro,
    Price: product.Price,
    Cover: product.Cover,
  }

  h.Response.Json(result)
}

func (h *ProductsHandler) Create(
  w http.ResponseWriter,
  r *http.Request,
) {
  h.Response = &api.ResponseHandler{
    Writer: w,
  }

  r.ParseForm()

  uid := r.Header.Get("uid")
  title := r.FormValue("title")
  intro := r.FormValue("intro")
  price, _ := strconv.ParseFloat(r.FormValue("price"), 64)
  cover := r.FormValue("cover")

  id := h.Repository.Create(
    uid,
    title,
    intro,
    price,
    cover,
  )

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)

  h.Response.Json(map[string]interface{}{
    "id": id,
  })
}

func (h *ProductsHandler) Update(
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

  title := r.FormValue("title")
  intro := r.FormValue("intro")
  price, _ := strconv.ParseFloat(r.FormValue("price"), 64)
  cover := r.FormValue("cover")

  h.Repository.Update(
    id,
    title,
    intro,
    price,
    cover,
  )

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)

  h.Response.Json(nil)
}
