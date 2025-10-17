package gfw

type LookupResponse struct {
  Records *LookupRecords `json:"records"`
}

type LookupRecords struct {
  A    *LookupItem `json:"a"`
  AAAA *LookupItem `json:"aaaa"`
}

type LookupItem struct {
  Response *LookupItemResponse `json:"response"`
}

type LookupItemResponse struct {
  Code   string              `json:"rCode"`
  Answer []*LookupItemAnswer `json:"answer"`
}

type LookupItemAnswer struct {
  Record *LookupItemRecord `json:"record"`
}

type LookupItemRecord struct {
  RecordType string `json:"recordType"`
  Raw        string `json:"raw"`
}
