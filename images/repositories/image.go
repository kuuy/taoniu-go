package repositories

import (
  "errors"

  "gorm.io/gorm"

  "github.com/rs/xid"

  . "taoniu.local/images/models"
)

type ImageRepository struct {
  db *gorm.DB
}

func NewImageRepository(db *gorm.DB) *ImageRepository {
  return &ImageRepository{
    db: db,
  }
}

func (r *ImageRepository) Listings() ([]Image, error) {
  offset := 0
  limit := 25

  var images []Image
  r.db.Select(
    "id",
    "title",
    "intro",
    "filehash",
    "ext",
  ).Order(
    "created_at desc",
  ).Offset(
    offset,
  ).Limit(
    limit,
  ).Find(
    &images,
  )

  return images, nil
}

func (r *ImageRepository) Get(id string) (Image, error) {
  var entity Image
  result := r.db.Where("id = ?", id).First(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return entity, result.Error
  }

  return entity, nil
}

func (r *ImageRepository) Save(
  width int64,
  height int64,
  mime string,
  size int64,
  filepath string,
  filename string,
  filehash string,
) Image {
  var entity Image
  result := r.db.Where("filehash", filehash).First(&entity)
  if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    entity = Image{
      ID:xid.New().String(),
      Title:"",
      Intro:"",
      Width:width,
      Height:height,
      Mime:mime,
      Size:size,
      Filepath:filepath,
      Filename:filename,
      Filehash:filehash,
    }
    r.db.Create(&entity)
  }

  return entity
}
