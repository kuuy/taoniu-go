package routers

import (
  "crypto/sha1"
  "encoding/json"
  "fmt"
  "image"
  "io"
  "math/rand"
  "net/http"
  "os"

  _ "image/gif"
  _ "image/jpeg"
  _ "image/png"

  "github.com/go-chi/chi/v5"
  "github.com/h2non/filetype"
  "github.com/rs/xid"

  pool "taoniu.local/images/common"
  "taoniu.local/images/repositories"
)

type ImageHandler struct {
  repository *repositories.ImageRepository
}

type Image struct {
  ID       string `json:"id"`
  Title    string `json:"title"`
  Intro    string `json:"intro"`
  Filepath string `json:"filepath"`
  Filename string `json:"filename"`
}

type ImageDetail struct {
  ID       string `json:"id"`
  Title    string `json:"title"`
  Intro    string `json:"intro"`
  Width    int    `json:"width"`
  Height   int    `json:"height"`
  Mime     string `json:"mime"`
  Size     uint64 `json:"size"`
  Filepath string `json:"filepath"`
  Filename string `json:"filename"`
  FileHash string `json:"file_hash"`
}

type ListImageResponse struct {
  Images []Image `json:"images"`
}

type UploadImageResponse struct {
  Filename string `json:"filename"`
}

const (
  MB              = 1 << 20
  MAX_UPLOAD_SIZE = 5 * MB
)

func NewImageRouter() http.Handler {
  db := pool.NewDB()
  repository := repositories.NewImageRepository(db)

  handler := ImageHandler{
    repository: repository,
  }

  r := chi.NewRouter()
  r.Get("/", handler.Listings)
  r.Post("/upload", handler.Upload)
  r.Get("/{id:[a-z0-9]{20}}.{ext}", handler.Display)

  return r
}

func (h *ImageHandler) Listings(w http.ResponseWriter, r *http.Request) {
  orders, err := h.repository.Listings()
  if err != nil {
    return
  }

  var response ListImageResponse
  for _, entity := range orders {
    var image Image
    image.ID = entity.ID

    response.Images = append(response.Images, image)
  }

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)

  jsonResponse, err := json.Marshal(response)
  if err != nil {
    return
  }

  w.Write(jsonResponse)
}

func (h *ImageHandler) Upload(
  w http.ResponseWriter,
  r *http.Request,
) {
  if err := r.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }

  file, _, err := r.FormFile("file")
  if err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }
  defer file.Close()

  head := make([]byte, 261)
  if _, err := file.Read(head); err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }

  kind, _ := filetype.Image(head)
  if kind == filetype.Unknown {
    http.Error(
      w,
      "not an image",
      http.StatusBadRequest,
    )
    return
  }

  if _, err := file.Seek(0, 0); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  homepath, err := os.UserHomeDir()
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  filepath := fmt.Sprintf(
    "%d/%d",
    rand.Intn(50),
    rand.Intn(50),
  )
  filename := fmt.Sprintf(
    "%s.%s",
    xid.New().String(),
    kind.Extension,
  )
  err = os.MkdirAll(
    fmt.Sprintf(
      "%s/.taoniu/images/%s",
      homepath,
      filepath,
    ),
    os.ModePerm,
  )
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  hash := sha1.New()
  dst, err := os.Create(
    fmt.Sprintf(
      "%s/.taoniu/images/%s/%s",
      homepath,
      filepath,
      filename,
    ),
  )
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  defer dst.Close()

  t := io.TeeReader(file, hash)

  _, err = io.Copy(dst, t)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  if _, err := file.Seek(0, 0); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  config, _, err := image.DecodeConfig(file)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  info, err := dst.Stat()
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  filehash := fmt.Sprintf("%x", hash.Sum(nil))

  entity := h.repository.Save(
    int64(config.Width),
    int64(config.Height),
    kind.MIME.Value,
    info.Size(),
    filepath,
    filename,
    filehash,
  )

  if entity.Filename != filename {
    os.Remove(
      fmt.Sprintf(
        "%s/.taoniu/images/%s/%s",
        homepath,
        filepath,
        filename,
      ),
    )
  }

  var response UploadImageResponse
  response.Filename = fmt.Sprintf(
    "%s.%s",
    entity.ID,
    kind.Extension,
  )

  jsonResponse, err := json.Marshal(response)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)

  w.Write(jsonResponse)
}

func (h *ImageHandler) Display(
  w http.ResponseWriter,
  r *http.Request,
) {
  id := chi.URLParam(r, "id")
  entity, err := h.repository.Get(id)
  if err != nil {
    http.Error(w, http.StatusText(404), http.StatusNotFound)
    return
  }

  homepath, err := os.UserHomeDir()
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  file, err := os.Open(
    fmt.Sprintf(
      "%s/.taoniu/images/%s/%s",
      homepath,
      entity.Filepath,
      entity.Filename,
    ),
  )
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  w.Header().Set("Content-Type", entity.Mime)
  http.ServeContent(w, r, "", entity.CreatedAt, file)
}
