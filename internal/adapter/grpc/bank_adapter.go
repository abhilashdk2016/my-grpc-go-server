package grpc

import (
	"context"
	"time"

	bank_proto "github.com/abhilashdk2016/my-grpc-proto/protogen/go/bank-proto"
	"google.golang.org/genproto/googleapis/type/date"
)

func (a *GrpcAdapter) GetCurrentBalance(ctx context.Context, req *bank_proto.CurrentBalanceRequest) (*bank_proto.CurrentBalanceResponse, error) {
	now := time.Now()
	bal := a.bankService.FindCurrentBalance(req.AccountNumber)
	return &bank_proto.CurrentBalanceResponse{
		Amount: bal,
		CurrentDate: &date.Date{
			Year:  int32(now.Year()),
			Month: int32(now.Month()),
			Day:   int32(now.Day()),
		},
	}, nil
}
