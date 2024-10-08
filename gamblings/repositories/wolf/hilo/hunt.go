package hilo

import (
  "context"
  "strconv"
  "strings"
  "time"

  "github.com/go-redis/redis/v8"
  "gorm.io/gorm"

  models "taoniu.local/gamblings/models/wolf/hilo"
)

type HuntRepository struct {
  Db  *gorm.DB
  Rdb *redis.Client
  Ctx context.Context
}

func (r *HuntRepository) Start() error {
  timestamp := time.Now().Unix()
  r.Rdb.ZAdd(r.Ctx, "wolf:hunts", &redis.Z{
    Score:  float64(timestamp),
    Member: "hilo",
  })
  return nil
}

func (r *HuntRepository) Gets(conditions map[string]interface{}) []*models.Hunt {
  var hunts []*models.Hunt

  query := r.Db.Select([]string{"hash", "number"})
  if _, ok := conditions["numbers"]; ok {
    query.Where("number IN ?", conditions["numbers"].([]float64))
  }
  if _, ok := conditions["ipart"]; ok {
    query.Where("ipart IN ?", conditions["ipart"].([]int))
  }
  if _, ok := conditions["dpart"]; ok {
    query.Where("dpart IN ?", conditions["dpart"].([]int))
  }
  if _, ok := conditions["side"]; ok {
    query.Where("side", conditions["side"].(int))
  }
  if _, ok := conditions["is_mirror"]; ok {
    query.Where("is_mirror", true)
  }
  if _, ok := conditions["is_repeate"]; ok {
    query.Where("is_repeate", true)
  }
  if _, ok := conditions["is_neighbor"]; ok {
    query.Where("is_neighbor", true)
  }
  if _, ok := conditions["opentime"]; ok {
    query.Where("updated_at > ?", conditions["opentime"].(time.Time))
  }
  query.Limit(50).Find(&hunts)

  return hunts
}

func (r *HuntRepository) Handing(hash string, number float64) error {
  parts := strings.Split(
    strconv.FormatFloat(number, 'f', -1, 64),
    ".",
  )
  if len(parts) == 1 {
    parts = append(parts, "0")
  }

  //ipart, _ := strconv.Atoi(parts[0])
  //dpart, _ := strconv.Atoi(parts[1])
  //
  //var hunt models.Hunt
  //result := r.Db.Where(
  //	"number=?",
  //	number,
  //).Take(&hunt)
  //if errors.Is(result.Error, gorm.ErrRecordNotFound) {
  //	side := r.Side(parts[0], parts[1])
  //	hunt = models.Hunt{
  //		ID:        xid.New().String(),
  //		Number:    number,
  //		Ipart:     uint8(ipart),
  //		Dpart:     uint8(dpart),
  //		Hash:      hash,
  //		Side:      side,
  //		IsMirror:  r.IsMirror(parts[0], parts[1]),
  //		IsRepeate: r.IsRepeate(parts[0], parts[1]),
  //	}
  //	if side != 0 {
  //		hunt.IsNeighbor = r.IsNeighbor(side, parts[0], parts[1])
  //	}
  //	r.Db.Apply(&hunt)
  //} else {
  //	hunt.Hash = hash
  //	r.Db.Model(&models.Hunt{ID: hunt.ID}).Updates(hunt)
  //}

  return nil
}

func (r *HuntRepository) Side(ipart string, dpart string) uint8 {
  var side uint8
  if ipart[0] < dpart[0] {
    side = 1
  } else if ipart[0] > dpart[0] {
    side = 2
  } else {
    side = 0
  }

  for i := 1; i < len(ipart); i++ {
    if side == 1 && ipart[i] <= ipart[i-1] {
      return 0
    }
    if side == 2 && ipart[i] >= ipart[i-1] {
      return 0
    }
  }

  if side == 1 && ipart[len(ipart)-1] >= dpart[0] {
    return 0
  }

  if side == 2 && ipart[len(ipart)-1] <= dpart[0] {
    return 0
  }

  for i := 1; i < len(dpart); i++ {
    if side == 1 && dpart[i] <= dpart[i-1] {
      return 0
    }
    if side == 2 && dpart[i] >= dpart[i-1] {
      return 0
    }
  }

  return side
}

func (r *HuntRepository) IsMirror(ipart string, dpart string) bool {
  if len(ipart) != len(dpart) {
    return false
  }
  for i := 0; i < len(ipart); i++ {
    if ipart[i] != dpart[len(ipart)-1-i] {
      return false
    }
  }
  return true
}

func (r *HuntRepository) IsRepeate(ipart string, dpart string) bool {
  if ipart != dpart {
    return false
  }
  return true
}

func (r *HuntRepository) IsNeighbor(side uint8, ipart string, dpart string) bool {
  for i := 1; i < len(ipart); i++ {
    if side == 1 && ipart[i] != ipart[i-1]+1 {
      return false
    }
    if side == 2 && ipart[i] != ipart[i-1]-1 {
      return false
    }
  }

  if side == 1 && ipart[len(ipart)-1] != dpart[0]+1 {
    return false
  }

  if side == 2 && ipart[len(ipart)-1] != dpart[0]-1 {
    return false
  }

  for i := 1; i < len(dpart); i++ {
    if side == 1 && dpart[i] != dpart[i-1]+1 {
      return false
    }
    if side == 2 && dpart[i] != dpart[i-1]-1 {
      return false
    }
  }

  return true
}
