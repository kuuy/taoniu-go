package tradings

type ScalpingPlacePayload struct {
  PlanId string `json:"plan_id"`
}

type ScalpingFlushPayload struct {
  ID string `json:"id"`
}

type TriggersPlacePayload struct {
  ID string `json:"id"`
}

type TriggersFlushPayload struct {
  ID string `json:"id"`
}
