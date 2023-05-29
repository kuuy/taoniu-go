package markets

import (
  "context"
  "github.com/go-redis/redis/v8"
  "google.golang.org/grpc"
  "google.golang.org/protobuf/types/known/timestamppb"
  "gorm.io/gorm"

  pb "taoniu.local/cryptos/grpc/binance/spot/markets/live"
  spotRepositories "taoniu.local/cryptos/repositories/binance/spot"
  repositories "taoniu.local/cryptos/repositories/binance/spot/markets"
)

type Live struct {
  pb.UnimplementedLiveServer
  Repository *repositories.LiveRepository
}

func NewLive(
  db *gorm.DB,
  rdb *redis.Client,
  ctx context.Context,
) *Live {
  repository := &repositories.LiveRepository{
    Db: db,
  }
  repository.TickersRepository = &spotRepositories.TickersRepository{
    Rdb: rdb,
    Ctx: ctx,
  }
  return &Live{
    Repository: repository,
  }
}

func (srv *Live) Pagenate(ctx context.Context, request *pb.PagenateRequest) (*pb.PagenateReply, error) {
  conditions := make(map[string]interface{})
  if request.Symbol != "" {
    conditions["symbol"] = request.Symbol
  }

  reply := &pb.PagenateReply{}
  reply.Total = srv.Repository.Count(conditions)
  data := srv.Repository.Listings(
    conditions,
    int(request.Current),
    int(request.PageSize),
  )
  for _, liveInfo := range data {
    reply.Data = append(reply.Data, &pb.LiveInfo{
      Symbol:    liveInfo.Symbol,
      Open:      liveInfo.Open,
      Price:     liveInfo.Price,
      High:      liveInfo.High,
      Low:       liveInfo.Low,
      Volume:    liveInfo.Volume,
      Quota:     liveInfo.Quota,
      Timestamp: timestamppb.New(liveInfo.Timestamp),
    })
  }
  return reply, nil
}

func (srv *Live) Register(s *grpc.Server) error {
  pb.RegisterLiveServer(s, srv)
  return nil
}
