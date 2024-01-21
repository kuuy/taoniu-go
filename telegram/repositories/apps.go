package repositories

import (
  "context"
  "errors"

  "github.com/fatih/color"
  "github.com/gotd/td/telegram"
  "github.com/gotd/td/telegram/auth"
  "github.com/gotd/td/telegram/updates"
  "github.com/gotd/td/telegram/updates/hook"
  "github.com/gotd/td/tg"
  "github.com/rs/xid"
  "go.uber.org/zap"
  "go.uber.org/zap/zapcore"
  "gorm.io/gorm"

  "taoniu.local/telegram/models"
)

type AppsRepository struct {
  Db  *gorm.DB
  Ctx context.Context
}

func (r *AppsRepository) NewSession(app *models.Apps) telegram.SessionStorage {
  return &AppsSession{
    Db:  r.Db,
    App: app,
  }
}

func (r *AppsRepository) Get(appID int) (app *models.Apps, err error) {
  err = r.Db.Where("app_id", appID).Take(&app).Error
  return
}

func (r *AppsRepository) Run(phone string, appID int, appHash string) (err error) {
  app, err := r.Get(appID)
  if errors.Is(err, gorm.ErrRecordNotFound) {
    app = &models.Apps{
      ID:      xid.New().String(),
      Phone:   phone,
      AppID:   appID,
      AppHash: appHash,
      Status:  1,
    }
    if err = r.Db.Create(&app).Error; err != nil {
      return
    }
  }

  log, _ := zap.NewDevelopment(zap.IncreaseLevel(zapcore.InfoLevel), zap.AddStacktrace(zapcore.FatalLevel))
  defer func() { _ = log.Sync() }()

  dispatcher := tg.NewUpdateDispatcher()
  gaps := updates.New(updates.Config{
    Handler: dispatcher,
    Logger:  log.Named("gaps"),
  })

  client := telegram.NewClient(appID, appHash, telegram.Options{
    Logger:         log,
    UpdateHandler:  gaps,
    SessionStorage: r.NewSession(app),
    Middlewares: []telegram.Middleware{
      hook.UpdateHook(gaps.Handle),
    },
  })

  dispatcher.OnNewChannelMessage(func(ctx context.Context, entities tg.Entities, update *tg.UpdateNewChannelMessage) error {
    log.Info("Channel message", zap.Any("message", update.Message))
    return nil
  })
  dispatcher.OnNewMessage(func(ctx context.Context, entities tg.Entities, update *tg.UpdateNewMessage) error {
    log.Info("Message", zap.Any("message", update.Message))
    return nil
  })

  err = client.Run(r.Ctx, func(ctx context.Context) error {
    flow := auth.NewFlow(AppsAuth{phone: phone}, auth.SendCodeOptions{})
    if err := client.Auth().IfNecessary(ctx, flow); err != nil {
      return err
    }
    user, err := client.Self(ctx)
    if err != nil {
      return err
    }
    color.Blue("Login successfully! ID: %d, Username: %s", user.ID, user.Username)

    return gaps.Run(ctx, client.API(), user.ID, updates.AuthOptions{
      OnStart: func(ctx context.Context) {
        log.Info("Gaps started")
      },
    })
    return nil
  })
  return
}
