package main

import (
	"context"

	pb "github.com/semi-technologies/contextionary/contextionary"
	schema "github.com/semi-technologies/contextionary/contextionary/schema"
)

func (s *server) IsWordPresent(ctx context.Context, word *pb.Word) (*pb.WordPresent, error) {
	i := s.combinedContextionary.WordToItemIndex(word.Word)
	return &pb.WordPresent{Present: i.IsPresent()}, nil
}

func (s *server) IsWordStopword(ctx context.Context, word *pb.Word) (*pb.WordStopword, error) {
	sw := s.stopwordDetector.IsStopWord(word.Word)
	return &pb.WordStopword{Stopword: sw}, nil
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

func (s *server) SafeGetSimilarWordsWithCertainty(ctx context.Context, params *pb.SimilarWordsParams) (*pb.SimilarWordsResults, error) {
	words := s.combinedContextionary.SafeGetSimilarWordsWithCertainty(params.Word, params.Certainty)
	return &pb.SimilarWordsResults{
		Words: pbWordsFromStrings(words),
	}, nil
}

func pbWordsFromStrings(input []string) []*pb.Word {
	output := make([]*pb.Word, len(input), len(input))
	for i, word := range input {
		output[i] = &pb.Word{Word: word}
	}

	return output
}
