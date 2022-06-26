package main

import (
	"os"
  "fmt"
	
  "github.com/urfave/cli/v2"
  "gorm.io/driver/postgres"
  "gorm.io/gorm"

  . "taoniu.local/groceries/models"
)

func main() {
  app := &cli.App{
    Name: "binance futures rules",
    Usage: "",
    Action: func(c *cli.Context) error {
      fmt.Println("error", c.Err)
      return nil
    },
    Commands: []*cli.Command{
      {
        Name: "migrate",
        Usage: "",
        Action: func(c *cli.Context) error {
          if err := migrate(); err != nil {
            return cli.NewExitError(err.Error(), 1)
          }
          return nil
        },
      },
    },
    Version: "0.0.0",
  }

  err := app.Run(os.Args)
  if err != nil {
    fmt.Println("app start fatal", err)
  }
}

func migrate() error {
  fmt.Println("process migrator")
  dsn := "host=localhost user=taoniu password=64EQJMn1O9JrZ2G4 dbname=taoniu     sslmode=disable"
  db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
  if err != nil {
    fmt.Println("database connect failed")
    return err
  }
  db.AutoMigrate(
    &Product{},
    &ProductBarcode{},
  )
  return nil
}

