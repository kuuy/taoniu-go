package tradings

type LaunchpadFlushPayload struct {
  ID string
}

type LaunchpadPlacePayload struct {
  ID string
}

type ScalpingPlacePayload struct {
  PlanID string `json:"plan_id"`
}

type ScalpingFlushPayload struct {
  ID string `json:"id"`
}

type TriggersFlushPayload struct {
  Symbol string
}

type TriggersPlacePayload struct {
  Symbol string
}
