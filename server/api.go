package main

import (
	"context"
	"fmt"
	"strings"
	"sync"

	pb "github.com/semi-technologies/contextionary/contextionary"
	core "github.com/semi-technologies/contextionary/contextionary/core"
	schema "github.com/semi-technologies/contextionary/contextionary/schema"
	"github.com/semi-technologies/contextionary/extensions"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *server) AddExtension(ctx context.Context, params *pb.ExtensionInput) (*pb.AddExtensionResult, error) {
	err := s.extensionStorer.Put(ctx, params.Concept, extensions.ExtensionInput{
		Definition: params.Definition,
		Weight:     params.Weight,
	})
	if err != nil {
		return nil, GrpcErrFromTyped(err)
	}

	return &pb.AddExtensionResult{}, nil
}

func (s *server) Meta(ctx context.Context, params *pb.MetaParams) (*pb.MetaOverview, error) {
	return &pb.MetaOverview{
		Version:   Version,
		WordCount: int64(s.combinedContextionary.GetNumberOfItems()),
	}, nil
}

func (s *server) IsWordPresent(ctx context.Context, word *pb.Word) (*pb.WordPresent, error) {
	asExtension, err := s.extensionLookerUpper.Lookup(word.Word)
	if err != nil {
		return nil, GrpcErrFromTyped(err)
	}

	if asExtension != nil {
		return &pb.WordPresent{Present: true}, nil // TODO: add note about extension
	}

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
	return res, GrpcErrFromTyped(err)
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

func (s *server) MultiVectorForWord(ctx context.Context, params *pb.WordList) (*pb.VectorList, error) {
	lock := &sync.Mutex{}
	out := make([]*pb.Vector, len(params.Words))
	var errors []error

	concurrent := s.config.MaximumBatchSize
	for i := 0; i < len(params.Words); i += concurrent {
		end := i + concurrent
		if end > len(params.Words) {
			end = len(params.Words)
		}

		batch := params.Words[i:end]

		var wg = &sync.WaitGroup{}
		for j, elem := range batch {
			wg.Add(1)
			go func(i, j int, word string) {
				defer wg.Done()
				word = strings.ToLower(word)
				vec, err := s.vectorizer.VectorForWord(word)
				if err != nil {
					lock.Lock()
					errors = append(errors, err)
					lock.Unlock()
					return
				}

				if vec == nil {
					lock.Lock()
					out[i+j] = &pb.Vector{}
					lock.Unlock()
					return
				}

				lock.Lock()
				out[i+j] = vectorToProto(vec.vector)
				lock.Unlock()

			}(i, j, elem.Word)
		}

		wg.Wait()
	}

	if len(errors) > 0 {
		return nil, joinErrors(errors)
	}

	return &pb.VectorList{
		Vectors: out,
	}, nil
}

func joinErrors(in []error) error {
	msgs := make([]string, len(in))
	for i, err := range in {
		msgs[i] = fmt.Sprintf("at pos %d: %v", i, err)
	}

	return fmt.Errorf("%s", strings.Join(msgs, ", "))
}

func (s *server) VectorForWord(ctx context.Context, params *pb.Word) (*pb.Vector, error) {
	wo, err := s.vectorizer.VectorForWord(params.Word)
	if err != nil {
		return nil, GrpcErrFromTyped(err)
	}

	if wo == nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("word %s is not in the contextionary", params.Word))
	}

	return vectorToProto(wo.vector), nil
}

func (s *server) VectorForCorpi(ctx context.Context, params *pb.Corpi) (*pb.Vector, error) {
	overrides := assembleOverrideMap(params.Overrides)
	vector, err := s.vectorizer.Corpi(params.Corpi, overrides)
	if err != nil {
		if err == ErrNoUsableWords {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return vectorToProto(vector), nil
}

func assembleOverrideMap(in []*pb.Override) map[string]string {
	if in == nil || len(in) == 0 {
		return nil
	}

	out := map[string]string{}
	for _, or := range in {
		out[or.Word] = or.Expression
	}

	return out
}

func (s *server) vectorForWords(words []string) (*core.Vector, error) {
	var vectors []core.Vector
	for _, word := range words {
		vector, err := s.vectorForWord(word)
		if err != nil {
			return nil, GrpcErrFromTyped(err)
		}

		if vector == nil {
			continue
		}

		vectors = append(vectors, *vector)
	}

	if len(vectors) == 0 {
		return nil, nil
	}

	return core.ComputeCentroid(vectors)
}

func (s *server) vectorForWord(word string) (*core.Vector, error) {
	wi := s.combinedContextionary.WordToItemIndex(word)
	if s.stopwordDetector.IsStopWord(word) {
		return nil, nil
	}

	if !wi.IsPresent() {
		return nil, nil
	}

	res, err := s.combinedContextionary.GetVectorForItemIndex(wi)
	return res, GrpcErrFromTyped(err)
}

func vectorToProto(in *core.Vector) *pb.Vector {
	a := in.ToArray()
	output := make([]*pb.VectorEntry, len(a), len(a))
	for i, entry := range a {
		output[i] = &pb.VectorEntry{Entry: entry}
	}

	source := make([]*pb.InputElement, len(in.Source))
	for i, s := range in.Source {
		source[i] = &pb.InputElement{
			Concept:    s.Concept,
			Occurrence: s.Occurrence,
			Weight:     float32(s.Weight),
		}
	}

	return &pb.Vector{Entries: output, Source: source}
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
		return nil, GrpcErrFromTyped(err)
	}
	words, err := s.itemIndexesToWords(ii)
	if err != nil {
		return nil, GrpcErrFromTyped(err)
	}

	return &pb.NearestWords{
		Distances: dist,
		Words:     words,
	}, nil
}

func (s *server) MultiNearestWordsByVector(ctx context.Context, params *pb.VectorNNParamsList) (*pb.NearestWordsList, error) {
	lock := &sync.Mutex{}
	out := make([]*pb.NearestWords, len(params.Params))
	var errors []error

	concurrent := s.config.MaximumBatchSize
	for i := 0; i < len(params.Params); i += concurrent {
		end := i + concurrent
		if end > len(params.Params) {
			end = len(params.Params)
		}

		batch := params.Params[i:end]

		var wg = &sync.WaitGroup{}
		for j, elem := range batch {
			wg.Add(1)
			go func(i, j int, elem *pb.VectorNNParams) {
				defer wg.Done()

				ii, dist, err := s.combinedContextionary.GetNnsByVector(vectorFromProto(elem.Vector), int(elem.N), int(elem.K))
				if err != nil {
					lock.Lock()
					errors = append(errors, GrpcErrFromTyped(err))
					lock.Unlock()
					return
				}

				words, err := s.itemIndexesToWords(ii)
				if err != nil {
					lock.Lock()
					errors = append(errors, GrpcErrFromTyped(err))
					lock.Unlock()
					return
				}

				vectors, err := s.itemIndexesToVectors(ii)
				if err != nil {
					lock.Lock()
					errors = append(errors, GrpcErrFromTyped(err))
					lock.Unlock()
					return
				}

				out[i+j] = &pb.NearestWords{
					Distances: dist,
					Words:     words,
					Vectors:   vectors,
				}
			}(i, j, elem)
		}

		wg.Wait()
	}

	if len(errors) > 0 {
		return nil, joinErrors(errors)
	}

	return &pb.NearestWordsList{
		Words: out,
	}, nil
}

func (s *server) itemIndexesToWords(in []core.ItemIndex) ([]string, error) {
	output := make([]string, len(in), len(in))
	for i, itemIndex := range in {
		w, err := s.combinedContextionary.ItemIndexToWord(itemIndex)
		if err != nil {
			return nil, GrpcErrFromTyped(err)
		}

		output[i] = w
	}

	return output, nil
}

func (s *server) itemIndexesToVectors(in []core.ItemIndex) (*pb.VectorList, error) {
	out := &pb.VectorList{
		Vectors: make([]*pb.Vector, len(in)),
	}

	for i, itemIndex := range in {
		vector, err := s.combinedContextionary.GetVectorForItemIndex(itemIndex)
		if err != nil {
			return nil, GrpcErrFromTyped(err)
		}

		out.Vectors[i] = vectorToProto(vector)
	}

	return out, nil
}
