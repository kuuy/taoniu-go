package daily

import (
  "context"
  "github.com/go-redis/redis/v8"
  "google.golang.org/grpc"
  "gorm.io/gorm"
  pb "taoniu.local/cryptos/grpc/binance/futures/indicators/daily/ranking"
  futuresRepositories "taoniu.local/cryptos/repositories/binance/futures"
  repositories "taoniu.local/cryptos/repositories/binance/futures/indicators/daily"
)

type Ranking struct {
  pb.UnimplementedRankingServer
  Repository *repositories.RankingRepository
}

func NewRanking(db *gorm.DB, rdb *redis.Client, ctx context.Context) *Ranking {
  repository := &repositories.RankingRepository{
    Rdb: rdb,
    Ctx: ctx,
  }
  repository.SymbolsRepository = &futuresRepositories.SymbolsRepository{
    Db: db,
  }
  return &Ranking{
    Repository: repository,
  }
}

func (srv *Ranking) Pagenate(ctx context.Context, request *pb.PagenateRequest) (*pb.PagenateReply, error) {
  reply := &pb.PagenateReply{}
  ranking := srv.Repository.Listings(
    request.Symbol,
    request.Fields,
    request.SortField,
    int(request.SortType),
    int(request.Current),
    int(request.PageSize),
  )
  reply.Total = int64(ranking.Total)
  reply.Data = ranking.Data
  return reply, nil
}

func (srv *Ranking) Register(s *grpc.Server) error {
  pb.RegisterRankingServer(s, srv)
  return nil
}
