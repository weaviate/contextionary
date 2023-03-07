package main

import (
	"fmt"
	"time"

	"github.com/weaviate/contextionary/compoundsplitting"

	"github.com/weaviate/contextionary/adapters/repos"
	core "github.com/weaviate/contextionary/contextionary/core"
	"github.com/weaviate/contextionary/contextionary/core/stopwords"
	"github.com/weaviate/contextionary/extensions"
)

func (s *server) init() error {
	s.logger.WithField("config", s.config).Debugf("starting up with this config")

	if err := s.loadRawContextionary(); err != nil {
		return err
	}

	swDetector, err := stopwords.NewFromFile(s.config.StopwordsFile)
	if err != nil {
		return err
	}
	s.stopwordDetector = swDetector

	if err := s.buildContextionary(); err != nil {
		return err
	}

	var er extensionRepo
	var extensionRetriever extensionLookerUpper

	// ExtensionsStorageMode == "weaviate" is now a default storage option
	er = repos.NewExtensionsRepo(s.logger, s.config, 1*time.Second)
	extensionRetriever = extensions.NewLookerUpper(er)

	compoundSplitter, err := s.initCompoundSplitter()
	if err != nil {
		return err
	}
	vectorizer, err := NewVectorizer(s.rawContextionary, s.stopwordDetector, s.config, s.logger,
		NewSplitter(), extensionRetriever, compoundSplitter)
	if err != nil {
		return err
	}

	s.vectorizer = vectorizer
	s.extensionStorer = extensions.NewStorer(s.vectorizer, er, s.logger)
	s.extensionLookerUpper = extensionRetriever

	return nil
}

func (s *server) loadRawContextionary() error {
	c, err := core.LoadVectorFromDisk(s.config.KNNFile, s.config.IDXFile)
	if err != nil {
		return fmt.Errorf("could not initialize (raw) contextionary: %v", err)
	}

	s.rawContextionary = c
	return nil
}

type stopwordDetector interface {
	IsStopWord(word string) bool
}

// any time the schema changes the contextionary needs to be rebuilt.
func (s *server) buildContextionary() error {
	s.combinedContextionary = s.rawContextionary
	return nil
}

func (s *server) initCompoundSplitter() (compoundSplitter, error) {
	if s.config.EnableCompundSplitting {
		dict, err := compoundsplitting.NewContextionaryDict(s.config.CompoundSplittingDictionaryFile)
		if err != nil {
			return nil, err
		}
		return compoundsplitting.NewSplitter(dict), nil
	} else {
		return compoundsplitting.NewNoopSplitter(), nil
	}
}

type extensionRepo interface {
	extensions.RetrieverRepo
	extensions.StorerRepo
}
