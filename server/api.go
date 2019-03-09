package main

import (
	"context"

	pb "github.com/creativesoftwarefdn/contextionary/contextionary"
)

func (s *server) IsWordPresent(ctx context.Context, word *pb.Word) (*pb.WordPresent, error) {
	i := s.combinedContextionary.WordToItemIndex(word.Word)
	return &pb.WordPresent{Present: i.IsPresent()}, nil
}

func (s *server) SchemaSearch(ctx context.Context, params *pb.SchemaSearchParams) (*pb.SchemaSearchResults, error) {

	return nil, nil
}
