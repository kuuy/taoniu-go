package tradings

type LaunchpadPlacePayload struct {
  ID string
}

type LaunchpadFlushPayload struct {
  ID string
}

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
