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

func (s *server) IsWordPresent(ctx context.Context, word *pb.Word) (*pb.WordPresent, error) {
	i := s.combinedContextionary.WordToItemIndex(word.Word)
	return &pb.WordPresent{Present: i.IsPresent()}, nil
}
