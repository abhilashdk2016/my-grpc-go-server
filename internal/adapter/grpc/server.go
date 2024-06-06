package grpc

import (
	"fmt"
	"log"
	"net"

	"github.com/abhilashdk2016/my-grpc-go-server/internal/port"
	bank_proto "github.com/abhilashdk2016/my-grpc-proto/protogen/go/bank-proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GrpcAdapter struct {
	grpcPort    int
	server      *grpc.Server
	bankService port.BankServicePort
	bank_proto.BankServiceServer
}

func NewGrpcAdapter(bankService port.BankServicePort, grpcPort int) *GrpcAdapter {
	return &GrpcAdapter{
		grpcPort:    grpcPort,
		bankService: bankService,
	}
}

func (a *GrpcAdapter) Run() {
	var err error

	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", a.grpcPort))

	if err != nil {
		log.Fatalf("Failed to listen on port %d: %v\n", a.grpcPort, err)
	}

	log.Printf("Server listening on port %d\n", a.grpcPort)

	grpcServer := grpc.NewServer()
	a.server = grpcServer
	reflection.Register(grpcServer)
	bank_proto.RegisterBankServiceServer(grpcServer, a)
	if err = grpcServer.Serve(listen); err != nil {
		log.Fatalf("Failed to serve gRPC on port %d: %v\n", a.grpcPort, err)
	}
}

func (a *GrpcAdapter) Stop() {
	a.server.Stop()
}
