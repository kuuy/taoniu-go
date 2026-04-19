package gfw

type ResolveResponse struct {
  Records *ResolveRecords `json:"records"`
}

type ResolveRecords struct {
  A    *ResolveData `json:"a"`
  AAAA *ResolveData `json:"aaaa"`
}

type ResolveData struct {
  Response *ResolveRecordResponse `json:"response"`
}

type ResolveRecordResponse struct {
  Code   string           `json:"rCode"`
  Answer []*ResolveAnswer `json:"answer"`
}

type ResolveAnswer struct {
  Record *ResolveRecord `json:"record"`
}

type ResolveRecord struct {
  RecordType string `json:"recordType"`
  Raw        string `json:"raw"`
}
