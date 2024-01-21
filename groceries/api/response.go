package api

import (
  "encoding/json"
  "net/http"
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
  Writer http.ResponseWriter
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
  json, err := json.Marshal(response)
  if err != nil {
    return
  }

  h.Writer.Write(json)
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
  json, err := json.Marshal(response)
  if err != nil {
    return
  }

  h.Writer.Write(json)
}

func (h *ResponseHandler) Error(status int, code int, message string) {
  h.Writer.Header().Set("Content-Type", "application/json")
  h.Writer.WriteHeader(status)

  response := errorResponse{}
  response.Success = false
  response.Code = code
  response.Message = message
  json, err := json.Marshal(response)
  if err != nil {
    return
  }

  h.Writer.Write(json)
}
