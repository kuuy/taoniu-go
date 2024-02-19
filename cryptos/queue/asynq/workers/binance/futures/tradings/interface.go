package tradings

type ScalpingPlacePayload struct {
  PlanID string `json:"plan_id"`
}

type ScalpingFlushPayload struct {
  ID string `json:"id"`
}

type TriggersPlacePayload struct {
  ID string
}

type TriggersFlushPayload struct {
  ID string
}