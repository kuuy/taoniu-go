package routers

import (
  "strconv"
  "encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	pool "taoniu.local/groceries/common"
	"taoniu.local/groceries/repositories"
)

type ProductHandler struct{
  repository *repositories.ProductRepository
}

type Product struct{
  ID string `json:"id"`
  Title string `json:"title"`
  Intro string `json:"intro"`
  Price float64 `json:"price"`
  Cover string `json:"cover"`
}

type ListProductResponse struct{
  Products []Product `json:"products"`
}

type ProductDetailResponse struct{
  ID string `json:"id"`
  Barcode string `json:"barcode"`
  Title string `json:"title"`
  Intro string `json:"intro"`
  Price float64 `json:"price"`
  Cover string `json:"cover"`
}

type ProductBarcodeResponse struct{
  ProductID string `json:"product_id"`
  Barcode string `json:"barcode"`
}

func NewProductRouter() http.Handler {
  db := pool.NewDB()
  repository := repositories.NewProductRepository(db)

  handler := ProductHandler{
    repository : repository,
  }

  r := chi.NewRouter()
  r.Get("/", handler.Listings)
  r.Get("/{id:[a-z0-9]{20}}", handler.Get)
  r.Post("/", handler.Create)
  r.Put("/{id:[a-z0-9]{20}}", handler.Update)
  r.Get("/barcode/{barcode}", handler.GetBarcode)

  return r
}

func (h *ProductHandler) Listings(w http.ResponseWriter, r *http.Request) {
  products, err := h.repository.Listings()
  if err != nil {
    return
  }

  var response ListProductResponse
  for _,entity := range(products) {
    var product Product
    product.ID = entity.ID
    product.Title = entity.Title
    product.Intro = entity.Intro
    product.Price = entity.Price
    product.Cover = entity.Cover

    response.Products = append(response.Products, product)
  }

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)

  jsonResponse, err := json.Marshal(response)
  if err != nil {
    return
  }

  w.Write(jsonResponse)
}

func (h *ProductHandler) Get(
  w http.ResponseWriter,
  r *http.Request,
) {
  id := chi.URLParam(r, "id")
  product, err := h.repository.Get(id)
  if err != nil {
    http.Error(w, http.StatusText(404), http.StatusNotFound)
    return
  }

  var response ProductDetailResponse
  response.ID = product.ID
  response.Title = product.Title
  response.Intro = product.Intro
  response.Price = product.Price
  response.Cover = product.Cover

  productBarcode, err := h.repository.GetProductBarcodeByProductID(
    product.ID,
  )
  if err != nil {
    response.Barcode = ""
  } else {
    response.Barcode = productBarcode.Barcode
  }

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)

  jsonResponse, err := json.Marshal(response)
  if err != nil {
    return
  }
 
  w.Write(jsonResponse)
}

func (h *ProductHandler) Create(
  w http.ResponseWriter,
  r *http.Request,
) {
  r.ParseForm()

  storeID := "000000"
  barcode := r.FormValue("barcode")
  title := r.FormValue("title")
  intro := r.FormValue("intro")
  price,_ := strconv.ParseFloat(r.FormValue("price"), 64)
  cover := r.FormValue("cover")

  h.repository.Create(
    storeID,
    barcode,
    title,
    intro,
    price,
    cover,
  )

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
}

func (h *ProductHandler) Update(
  w http.ResponseWriter,
  r *http.Request,
) {
  r.ParseForm()

  storeID := "000000"

  id := chi.URLParam(r, "id")
  _, err := h.repository.Get(id)
  if err != nil {
    http.Error(w, http.StatusText(404), http.StatusNotFound)
    return
  }

  title := r.FormValue("title")
  barcode := r.FormValue("barcode")
  intro := r.FormValue("intro")
  price,_ := strconv.ParseFloat(r.FormValue("price"), 64)
  cover := r.FormValue("cover")

  h.repository.Update(
    id,
    storeID,
    barcode,
    title,
    intro,
    price,
    cover,
  )

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
}

func (h *ProductHandler) GetBarcode(
  w http.ResponseWriter,
  r *http.Request,
) {
  storeID := "000000"
  barcode := chi.URLParam(r, "barcode")
  productBarcode, err := h.repository.GetProductBarcode(storeID, barcode)
  if err != nil {
    http.Error(w, http.StatusText(404), http.StatusNotFound)
    return
  }

  var response ProductBarcodeResponse
  response.ProductID = productBarcode.ProductID
  response.Barcode = productBarcode.Barcode

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)

  jsonResponse, err := json.Marshal(response)
  if err != nil {
    return
  }
 
  w.Write(jsonResponse)
}
