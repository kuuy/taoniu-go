package repositories

import (
  "context"
  "log"

  "github.com/gotd/td/telegram"
  "gorm.io/gorm"
)

type BotsRepository struct {
  Db     *gorm.DB
  Ctx    context.Context
  Client *telegram.Client
}

func (r *BotsRepository) Auth(token string) error {
  log.Println("token", token)
  r.Client.Run(r.Ctx, func(ctx context.Context) error {
    // Grab token from @BotFather.
    if _, err := r.Client.Auth().Bot(ctx, token); err != nil {
      return err
    }
    state, err := r.Client.API().UpdatesGetState(ctx)
    if err != nil {
      return err
    }
    log.Println("state", state)
    return nil
  })

  return nil
}
