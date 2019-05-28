package main

import (
	"context"
	"fmt"

	pb "github.com/semi-technologies/contextionary/contextionary"
	core "github.com/semi-technologies/contextionary/contextionary/core"
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

func (s *server) VectorForWord(ctx context.Context, params *pb.Word) (*pb.Vector, error) {
	i := s.combinedContextionary.WordToItemIndex(params.Word)
	if !i.IsPresent() {
		return nil, fmt.Errorf("word %s is not in the contextionary", params.Word)
	}

	v, err := s.combinedContextionary.GetVectorForItemIndex(i)
	if err != nil {
		return nil, err
	}

	return vectorToProto(v), nil
}

func vectorToProto(in *core.Vector) *pb.Vector {
	a := in.ToArray()
	output := make([]*pb.VectorEntry, len(a), len(a))
	for i, entry := range a {
		output[i] = &pb.VectorEntry{Entry: entry}
	}

	return &pb.Vector{Entries: output}
}

func vectorFromProto(in *pb.Vector) core.Vector {
	asFloats := make([]float32, len(in.Entries), len(in.Entries))
	for i, entry := range in.Entries {
		asFloats[i] = entry.Entry
	}

	return core.NewVector(asFloats)
}

func (s *server) NearestWordsByVector(ctx context.Context, params *pb.VectorNNParams) (*pb.NearestWords, error) {

	ii, dist, err := s.combinedContextionary.GetNnsByVector(vectorFromProto(params.Vector), int(params.N), int(params.K))
	if err != nil {
		return nil, err
	}
	words, err := s.itemIndexesToWords(ii)
	if err != nil {
		return nil, err
	}

	return &pb.NearestWords{
		Distances: dist,
		Words:     words,
	}, nil
}

func (s *server) itemIndexesToWords(in []core.ItemIndex) ([]string, error) {
	output := make([]string, len(in), len(in))
	for i, itemIndex := range in {
		w, err := s.combinedContextionary.ItemIndexToWord(itemIndex)
		if err != nil {
			return nil, err
		}

		output[i] = w
	}

	return output, nil
}
