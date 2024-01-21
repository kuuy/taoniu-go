package repositories

import (
  "context"
  "errors"
  "gorm.io/gorm"
  "strings"
  "time"

  "github.com/gotd/td/telegram/auth"
  "github.com/gotd/td/tg"
  "github.com/tcnksm/go-input"

  "taoniu.local/telegram/models"
)

type AppsAuth struct {
  noSignUp
  phone string
}

type AppsSession struct {
  Db  *gorm.DB
  App *models.Apps
}

type noSignUp struct{}

func (s noSignUp) SignUp(_ context.Context) (auth.UserInfo, error) {
  return auth.UserInfo{}, errors.New("searchx don't support sign up Telegram account")
}

func (s noSignUp) AcceptTermsOfService(_ context.Context, tos tg.HelpTermsOfService) error {
  return &auth.SignUpRequired{TermsOfService: tos}
}

func (a AppsAuth) Phone(_ context.Context) (string, error) {
  return a.phone, nil
}

func (a AppsAuth) Password(_ context.Context) (string, error) {
  pwd, err := input.DefaultUI().Ask("Enter 2FA Password:", &input.Options{
    Required: true,
    Loop:     true,
  })
  if err != nil {
    return "", err
  }
  return strings.TrimSpace(pwd), nil
}

func (a AppsAuth) Code(_ context.Context, _ *tg.AuthSentCode) (string, error) {
  code, err := input.DefaultUI().Ask("Enter Code:", &input.Options{
    Required: true,
    Loop:     true,
  })
  if err != nil {
    return "", err
  }
  return strings.TrimSpace(code), nil
}

func (s *AppsSession) LoadSession(ctx context.Context) (session []byte, err error) {
  session = []byte(s.App.Session)
  return
}

func (s *AppsSession) StoreSession(ctx context.Context, data []byte) (err error) {
  s.Db.Model(&s.App).Updates(map[string]interface{}{
    "session":   string(data[:]),
    "timestamp": time.Now().UnixMicro(),
  })
  return
}
