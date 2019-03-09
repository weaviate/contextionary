package main

import (
	"fmt"

	core "github.com/creativesoftwarefdn/contextionary/contextionary/core"
	schemac "github.com/creativesoftwarefdn/contextionary/contextionary/schema"
	"github.com/creativesoftwarefdn/weaviate/database/schema"
	"github.com/creativesoftwarefdn/weaviate/models"
)

func (s *server) init() error {
	s.logger.WithField("config", s.config).Debugf("starting up with this config")

	s.schema = emptySchema()
	if err := s.loadRawContextionary(); err != nil {
		return err
	}

	if err := s.buildContextionary(); err != nil {
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

// any time the schema changes the contextionary needs to be rebuilt.
func (s *server) buildContextionary() error {
	schemaContextionary, err := schemac.BuildInMemoryContextionaryFromSchema(s.schema, &s.rawContextionary)
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
		Things:  &models.SemanticSchema{},
	}
}
