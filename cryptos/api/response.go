package api

import (
  "encoding/json"
  "net/http"
  "taoniu.local/cryptos/repositories"
)

type jsonResponse struct {
  Success bool        `json:"success"`
  Data    interface{} `json:"data"`
}

type pagenateResponse struct {
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
  JweRepository *repositories.JweRepository
  Writer        http.ResponseWriter
}

func (h *ResponseHandler) Out(data interface{}) {
  h.Writer.Header().Set("Content-Type", "application/json")
  //h.Writer.WriteHeader(http.StatusOK)

  json, err := json.Marshal(data)
  if err != nil {
    return
  }

  h.Writer.Write(json)
}

func (h *ResponseHandler) Json(data interface{}) {
  h.Writer.Header().Set("Content-Type", "application/json")
  //h.Writer.WriteHeader(http.StatusOK)

  response := jsonResponse{}
  response.Success = true
  response.Data = data
  payload, err := json.Marshal(response)
  if err != nil {
    return
  }

  jweCompact, _ := h.JweRepository.Encrypt(payload)
  h.Writer.Write([]byte(jweCompact))
}

func (h *ResponseHandler) Pagenate(
  data interface{},
  total int64,
  current int,
  pageSize int,
) {
  h.Writer.Header().Set("Content-Type", "application/json")
  //h.Writer.WriteHeader(http.StatusOK)

  response := pagenateResponse{}
  response.Success = true
  response.Data = data
  response.Total = total
  response.PageSize = pageSize
  response.Current = current
  payload, err := json.Marshal(response)
  if err != nil {
    return
  }

  jweCompact, _ := h.JweRepository.Encrypt(payload)
  h.Writer.Write([]byte(jweCompact))
}

func (h *ResponseHandler) Error(status int, code int, message string) {
  h.Writer.Header().Set("Content-Type", "application/json")
  h.Writer.WriteHeader(status)

  response := errorResponse{}
  response.Success = false
  response.Code = code
  response.Message = message
  payload, err := json.Marshal(response)
  if err != nil {
    return
  }

  jweCompact, _ := h.JweRepository.Encrypt(payload)
  h.Writer.Write([]byte(jweCompact))
}
