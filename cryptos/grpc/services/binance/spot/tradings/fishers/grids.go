package fishers

import (
  "context"
  "google.golang.org/grpc"
  "google.golang.org/protobuf/types/known/timestamppb"
  "gorm.io/gorm"
  pb "taoniu.local/cryptos/grpc/binance/spot/tradings/fishers/grids"
  repositories "taoniu.local/cryptos/repositories/binance/spot/tradings/fishers"
)

type Grids struct {
  pb.UnimplementedGridsServer
  Repository *repositories.GridsRepository
}

func NewGrids(db *gorm.DB) *Grids {
  return &Grids{
    Repository: &repositories.GridsRepository{
      Db: db,
    },
  }
}

func (srv *Grids) Pagenate(ctx context.Context, request *pb.PagenateRequest) (*pb.PagenateReply, error) {
  reply := &pb.PagenateReply{}
  reply.Total = srv.Repository.Count()
  grids := srv.Repository.Listings(int(request.Page), int(request.PageSize))
  for _, grid := range grids {
    reply.Data = append(reply.Data, &pb.GridInfo{
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

func (srv *Grids) Register(s *grpc.Server) error {
  pb.RegisterGridsServer(s, srv)
  return nil
}
