package futures

import (
  "context"

  "google.golang.org/grpc"
  "google.golang.org/protobuf/types/known/timestamppb"
  "gorm.io/gorm"

  pb "taoniu.local/cryptos/grpc/binance/futures/plans"
  repositories "taoniu.local/cryptos/repositories/binance/futures"
)

type Plans struct {
  pb.UnimplementedPlansServer
  Repository *repositories.PlansRepository
}

func NewPlans(db *gorm.DB) *Plans {
  repository := &repositories.PlansRepository{
    Db: db,
  }
  return &Plans{
    Repository: repository,
  }
}

func (srv *Plans) Pagenate(ctx context.Context, request *pb.PagenateRequest) (*pb.PagenateReply, error) {
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
  for _, plan := range data {
    reply.Data = append(reply.Data, &pb.PlanInfo{
      Symbol:    plan.Symbol,
      Side:      uint32(plan.Side),
      Price:     float32(plan.Price),
      Quantity:  float32(plan.Quantity),
      Amount:    float32(plan.Amount),
      CreatedAt: timestamppb.New(plan.CreatedAt),
    })
  }
  return reply, nil
}

func (srv *Plans) Register(s *grpc.Server) error {
  pb.RegisterPlansServer(s, srv)
  return nil
}
