package main

import (
	"fmt"

	core "github.com/semi-technologies/contextionary/contextionary/core"
	"github.com/semi-technologies/contextionary/contextionary/core/stopwords"
	schemac "github.com/semi-technologies/contextionary/contextionary/schema"
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

	if err := s.buildContextionary(swDetector); err != nil {
		return err
	}

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
func (s *server) buildContextionary(swDetector stopwordDetector) error {
	schemaContextionary, err := schemac.BuildInMemoryContextionaryFromSchema(s.schema, &s.rawContextionary, swDetector)
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
		Actions: &models.SemanticSchema{},
		Things: &models.SemanticSchema{
			Classes: []*models.SemanticSchemaClass{
				{
					Class: "City",
					Properties: []*models.SemanticSchemaClassProperty{{
						Name:     "name",
						DataType: []string{"string"},
					}},
				},
				{
					Class: "Village",
					Properties: []*models.SemanticSchemaClassProperty{{
						Name:     "name",
						DataType: []string{"string"},
					}},
				},
			},
		},
	}
}
