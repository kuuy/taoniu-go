package repositories

import (
  "errors"

  "gorm.io/gorm"

  "github.com/rs/xid"

  . "taoniu.local/groceries/models"
)

type ProductRepository struct {
  db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
  return &ProductRepository{
    db: db,
  }
}

func (r *ProductRepository) Listings() ([]Product, error) {
  offset := 0
  limit := 25

  var products []Product
  r.db.Select(
    "id",
    "title",
    "intro",
    "price",
    "cover",
  ).Order(
    "created_at desc",
  ).Offset(
    offset,
  ).Limit(
    limit,
  ).Find(
    &products,
  )

  return products, nil
}

func (r *ProductRepository) Get(
  id string,
) (Product, error) {
  var product Product
  result := r.db.Where("id = ?", id).First(&product)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return product, result.Error
  }

  return product, nil
}

func (r *ProductRepository) Create(
  storeID string,
  barcode string,
  title string,
  intro string,
  price float64,
  cover string,
) {
  var ID = xid.New().String()
  r.db.Create(&Product{
    ID:ID,
    StoreID:storeID,
    Title:title,
    Intro:intro,
    Price:price,
    Cover:cover,
  })

  if (len(barcode) > 0) {
    r.SaveProductBarcode(
      storeID,
      ID,
      barcode,
    )
  }
}

func (r *ProductRepository) Update(
  id string,
  storeID string,
  barcode string,
  title string,
  intro string,
  price float64,
  cover string,
) error {
  var product Product
  result := r.db.Where("id = ?", id).First(&product)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return result.Error
  }

  product.Title = title
  product.Intro = intro
  product.Price = price
  product.Cover = cover
  r.db.Where("id = ?", id).Updates(product)

  if (len(barcode) > 0) {
    r.SaveProductBarcode(
      product.StoreID, 
      product.ID, 
      barcode,
    )
  }

  return nil
}

func (r *ProductRepository) GetProductBarcode(
  storeID string,
  barcode string,
) (ProductBarcode, error) {
  var productBarcode ProductBarcode
  result := r.db.Where(
    "store_id = ? AND barcode = ?",
    storeID,
    barcode,
  ).First(&productBarcode)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return productBarcode, result.Error
  }

  return productBarcode, nil
}

func (r *ProductRepository) GetProductBarcodeByProductID(
  productID string,
) (ProductBarcode, error) {
  var productBarcode ProductBarcode
  result := r.db.Where(
    "product_id = ?",
    productID,
  ).First(&productBarcode)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return productBarcode, result.Error
  }

  return productBarcode, nil
}

func (r *ProductRepository) SaveProductBarcode(
  storeID string,
  productID string,
  barcode string,
) {
  var productBarcode ProductBarcode
  result := r.db.Where(
    "store_id = ? AND barcode = ?",
    storeID,
    barcode,
  ).First(&productBarcode)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    r.db.Delete(&ProductBarcode{ProductID:productID})
    r.db.Create(&ProductBarcode{
      ID:xid.New().String(),
      StoreID:storeID,
      ProductID:productID,
      Barcode:barcode,
    })
  }
}
