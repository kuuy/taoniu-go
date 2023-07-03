package daily

import (
  "context"
  "fmt"
  "github.com/go-redis/redis/v8"
  "sort"
  "strconv"
  "strings"
  "time"
)

type RankingRepository struct {
  Rdb               *redis.Client
  Ctx               context.Context
  SymbolsRepository SymbolsRepository
}

type SymbolsRepository interface {
  Symbols() []string
}

type RankingScore struct {
  Symbol string
  Value  float64
  Data   []string
}

type RankingResult struct {
  Total int
  Data  []string
}

func (r *RankingRepository) Listings(
  symbol string,
  fields []string,
  sortField string,
  sortType int,
  current int,
  pageSize int,
) *RankingResult {
  var script = redis.NewScript(`
	local hmget = function (key)
		local hash = {}
		local data = redis.call('HMGET', key, unpack(ARGV))
		for i = 1, #ARGV do
			hash[i] = data[i]
		end
		return hash
	end
	local data = {}
	for i = 1, #KEYS do
		local key = 'binance:futures:indicators:1d:' .. KEYS[i]
		if redis.call('EXISTS', key) == 0 then
			data[i] = false
		else
			data[i] = hmget(key)
		end
	end
	return data
  `)

  var symbols []string
  if symbol == "" {
    symbols = r.SymbolsRepository.Symbols()
  } else {
    symbols = append(symbols, symbol)
  }

  sortIdx := -1
  day := time.Now().Format("0102")

  var keys []string
  for _, symbol := range symbols {
    keys = append(keys, fmt.Sprintf("%s:%s", symbol, day))
  }

  var args []interface{}
  for i, field := range fields {
    if field == sortField {
      sortIdx = i
    }
    args = append(args, field)
  }

  ranking := &RankingResult{}

  if sortIdx == -1 {
    return ranking
  }

  result, _ := script.Run(r.Ctx, r.Rdb, keys, args...).Result()

  var scores []*RankingScore
  for i := 0; i < len(symbols); i++ {
    item := result.([]interface{})[i]
    if item == nil {
      continue
    }
    if item.([]interface{})[sortIdx] == nil {
      continue
    }
    data := make([]string, len(fields))
    for j := 0; j < len(fields); j++ {
      if item.([]interface{})[j] == nil {
        continue
      }
      data[j] = fmt.Sprintf("%v", item.([]interface{})[j])
    }
    score, _ := strconv.ParseFloat(
      fmt.Sprintf("%v", item.([]interface{})[sortIdx]),
      16,
    )
    scores = append(scores, &RankingScore{
      symbols[i],
      score,
      data,
    })
  }

  if len(scores) == 0 {
    return ranking
  }

  sort.SliceStable(scores, func(i, j int) bool {
    if sortType == -1 {
      return scores[i].Value > scores[j].Value
    } else if sortType == 1 {
      return scores[i].Value < scores[j].Value
    }
    return true
  })

  offset := (current - 1) * pageSize
  endPos := offset + pageSize
  if endPos > len(scores) {
    endPos = len(scores)
  }

  ranking.Total = len(scores)
  for _, score := range scores[offset:endPos] {
    ranking.Data = append(ranking.Data, strings.Join(
      append([]string{score.Symbol}, score.Data...),
      ",",
    ))
  }

  return ranking
}
