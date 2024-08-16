package v1

type LoginRequest struct {
  Email    string `json:"email"`
  Password string `json:"password"`
}

type LoginResponse struct {
  AccessToken  string `json:"access_token"`
  RefreshToken string `json:"refresh_token"`
}

type RefreshTokenRequest struct {
  RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
  AccessToken string `json:"access_token"`
}
