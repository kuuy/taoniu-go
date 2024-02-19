package tradings

type ScalpingPlacePayload struct {
  PlanID string `json:"plan_id"`
}

type ScalpingFlushPayload struct {
  ID string `json:"id"`
}

type TriggersFlushPayload struct {
  ID string
}

type TriggersPlacePayload struct {
  ID string
}
