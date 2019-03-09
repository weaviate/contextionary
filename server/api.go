package main

import (
	"context"

	pb "github.com/creativesoftwarefdn/contextionary/contextionary"
)

func (s *server) GetHello(ctx context.Context, in *pb.Void) (*pb.Message, error) {

	return &pb.Message{
		Message: "hello world",
	}, nil
}
