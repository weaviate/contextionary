package main

import (
	"context"
	"fmt"
	"net"
	"os"

	pb "github.com/creativesoftwarefdn/contextionary/contextionary"
	grpc "google.golang.org/grpc"
)

type server struct {
}

func (s *server) GetHello(ctx context.Context, in *pb.Void) (*pb.Message, error) {

	return &pb.Message{
		Message: "hello world",
	}, nil
}

func main() {
	grpcServer := grpc.NewServer()
	pb.RegisterContextionaryServer(grpcServer, &server{})
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 9999))
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't listen on port: %s", err)
		os.Exit(1)
	}

	grpcServer.Serve(lis)
}
