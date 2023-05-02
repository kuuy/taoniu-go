package fishers

import (
  "context"
  "google.golang.org/grpc"
  "gorm.io/gorm"

  pb "taoniu.local/cryptos/grpc/binance/spot/analysis/tradings/fishers/chart"
  repositories "taoniu.local/cryptos/repositories/binance/spot/analysis/tradings/fishers"
)

type Chart struct {
  pb.UnimplementedChartServer
  Repository *repositories.ChartRepository
}

func NewChart(db *gorm.DB) *Chart {
  return &Chart{
    Repository: &repositories.ChartRepository{
      Db: db,
    },
  }
}

func (srv *Chart) Series(ctx context.Context, request *pb.SeriesRequest) (*pb.SeriesReply, error) {
  reply := &pb.SeriesReply{}
  reply.Series = srv.Repository.Series(int(request.Limit))
  return reply, nil
}

func (srv *Chart) Register(s *grpc.Server) error {
  pb.RegisterChartServer(s, srv)
  return nil
}
