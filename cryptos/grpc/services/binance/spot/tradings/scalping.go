package tradings

import (
  "context"
  "google.golang.org/grpc"
  "google.golang.org/protobuf/types/known/timestamppb"
  "gorm.io/gorm"
  pb "taoniu.local/cryptos/grpc/binance/spot/tradings/scalping"
  repositories "taoniu.local/cryptos/repositories/binance/spot/tradings"
)

type Scalping struct {
  pb.UnimplementedScalpingServer
  Repository *repositories.ScalpingRepository
}

func NewScalping(db *gorm.DB) *Scalping {
  return &Scalping{
    Repository: &repositories.ScalpingRepository{
      Db: db,
    },
  }
}

func (srv *Scalping) Pagenate(ctx context.Context, request *pb.PagenateRequest) (*pb.PagenateReply, error) {
  conditions := make(map[string]interface{})
  if request.Symbol != "" {
    conditions["symbol"] = request.Symbol
  }
  if len(request.Status) > 0 {
    var status []int
    for _, item := range request.Status {
      status = append(status, int(item))
    }
    conditions["status"] = status
  }

  reply := &pb.PagenateReply{}
  reply.Total = srv.Repository.Count(conditions)
  grids := srv.Repository.Listings(
    conditions,
    int(request.Page),
    int(request.PageSize),
  )
  for _, grid := range grids {
    reply.Data = append(reply.Data, &pb.ScalpingInfo{
      Id:           grid.ID,
      Symbol:       grid.Symbol,
      BuyPrice:     float32(grid.BuyPrice),
      BuyQuantity:  float32(grid.BuyQuantity),
      SellPrice:    float32(grid.SellPrice),
      SellQuantity: float32(grid.SellQuantity),
      Status:       int32(grid.Status),
      CreatedAt:    timestamppb.New(grid.CreatedAt),
      UpdatedAt:    timestamppb.New(grid.UpdatedAt),
    })
  }
  return reply, nil
}

func (srv *Scalping) Register(s *grpc.Server) error {
  pb.RegisterScalpingServer(s, srv)
  return nil
}
