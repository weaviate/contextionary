package main

import (
	"fmt"
	"time"

	"github.com/semi-technologies/contextionary/compoundsplitting"

	"github.com/coreos/etcd/clientv3"
	"github.com/semi-technologies/contextionary/adapters/repos"
	core "github.com/semi-technologies/contextionary/contextionary/core"
	"github.com/semi-technologies/contextionary/contextionary/core/stopwords"
	schemac "github.com/semi-technologies/contextionary/contextionary/schema"
	"github.com/semi-technologies/contextionary/extensions"
	"github.com/semi-technologies/weaviate/entities/models"
	"github.com/semi-technologies/weaviate/entities/schema"
)

func (s *server) init() error {
	s.logger.WithField("config", s.config).Debugf("starting up with this config")

	s.schema = emptySchema()
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

	if s.config.ExtensionsStorageMode == "weaviate" {
		er = repos.NewExtensionsRepo(s.logger, s.config, 1*time.Second)
		extensionRetriever = extensions.NewLookerUpper(er)
	} else {
		etcdClient, err := clientv3.New(clientv3.Config{
			Endpoints: []string{s.config.SchemaProviderURL},
		})
		if err != nil {
			return err
		}
		er = repos.NewEtcdExtensionRepo(etcdClient, s.logger, s.config)
		extensionRetriever = extensions.NewLookerUpper(er)
	}

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
	schemaContextionary, err := schemac.BuildInMemoryContextionaryFromSchema(s.schema, &s.rawContextionary, s.stopwordDetector)
	if err != nil {
		return fmt.Errorf("could not build in-memory contextionary from schema; %v", err)
	}

	// Combine contextionaries
	contextionaries := []core.Contextionary{s.rawContextionary, *schemaContextionary}
	combined, err := core.CombineVectorIndices(contextionaries)
	if err != nil {
		return fmt.Errorf("could not combine the contextionary database with the in-memory generated contextionary: %v", err)
	}

	// messaging.InfoMessage("Contextionary extended with names in the schema")

	s.combinedContextionary = combined

	return nil
}

func emptySchema() schema.Schema {
	return schema.Schema{
		Actions: &models.Schema{},
		Things:  &models.Schema{},
	}
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
