package api

import (
  "encoding/json"
  "net/http"
  "taoniu.local/cryptos/common"
)

type jsonResponse struct {
  Success bool        `json:"success"`
  Data    interface{} `json:"data"`
}

type paginateResponse struct {
  Success  bool        `json:"success"`
  Data     interface{} `json:"data"`
  Total    int64       `json:"total"`
  Current  int         `json:"current"`
  PageSize int         `json:"page_size"`
}

type errorResponse struct {
  Success bool   `json:"success"`
  Code    int    `json:"code"`
  Message string `json:"message"`
}

type ResponseHandler struct {
  Jwe *common.Jwe
}

func (h *ResponseHandler) Out(w http.ResponseWriter, data interface{}) {
  w.Header().Set("Content-Type", "application/json")

  jsonBytes, err := json.Marshal(data)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  w.Write(jsonBytes)
}

func (h *ResponseHandler) Json(w http.ResponseWriter, data interface{}) {
  w.Header().Set("Content-Type", "application/json")

  response := jsonResponse{
    Success: true,
    Data:    data,
  }
  payload, err := json.Marshal(response)
  if err != nil {
    h.Error(w, http.StatusInternalServerError, 5000, "json marshal error")
    return
  }

  jweCompact, err := h.Jwe.Encrypt(payload)
  if err != nil {
    h.Error(w, http.StatusInternalServerError, 5001, "encrypt error")
    return
  }
  w.Write([]byte(jweCompact))
}

func (h *ResponseHandler) Paginate(
  w http.ResponseWriter,
  data interface{},
  total int64,
  current int,
  pageSize int,
) {
  w.Header().Set("Content-Type", "application/json")

  response := paginateResponse{
    Success:  true,
    Data:     data,
    Total:    total,
    Current:  current,
    PageSize: pageSize,
  }
  payload, err := json.Marshal(response)
  if err != nil {
    h.Error(w, http.StatusInternalServerError, 5000, "json marshal error")
    return
  }

  jweCompact, err := h.Jwe.Encrypt(payload)
  if err != nil {
    h.Error(w, http.StatusInternalServerError, 5001, "encrypt error")
    return
  }
  w.Write([]byte(jweCompact))
}

func (h *ResponseHandler) Error(w http.ResponseWriter, status int, code int, message string) {
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(status)

  response := errorResponse{
    Success: false,
    Code:    code,
    Message: message,
  }
  payload, err := json.Marshal(response)
  if err != nil {
    http.Error(w, "internal error", http.StatusInternalServerError)
    return
  }

  jweCompact, err := h.Jwe.Encrypt(payload)
  if err != nil {
    http.Error(w, "encrypt error", http.StatusInternalServerError)
    return
  }
  w.Write([]byte(jweCompact))
}
