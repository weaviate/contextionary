package main

import (
	"context"

	pb "github.com/creativesoftwarefdn/contextionary/contextionary"
	schema "github.com/creativesoftwarefdn/contextionary/contextionary/schema"
)

func (s *server) IsWordPresent(ctx context.Context, word *pb.Word) (*pb.WordPresent, error) {
	i := s.combinedContextionary.WordToItemIndex(word.Word)
	return &pb.WordPresent{Present: i.IsPresent()}, nil
}

func (s *server) SchemaSearch(ctx context.Context, params *pb.SchemaSearchParams) (*pb.SchemaSearchResults, error) {

	s.logger.WithField("params", params).Info()
	c := schema.New(s.combinedContextionary)
	res, err := c.SchemaSearch(params)
	s.logger.
		WithField("res", res).
		WithField("err", err).Info()
	return res, err
}
