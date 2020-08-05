package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/semi-technologies/contextionary/compoundsplitting"
	"os"

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

	go s.watchForSchemaChanges()

	er := repos.NewEtcdExtensionRepo(s.etcdClient, s.logger, s.config)
	extensionRetriever := extensions.NewLookerUpper(er)

	dict, err := compoundsplitting.NewContextionaryDictFromFile(s.config.CompoundSplittingDictionaryFile)
	if err != nil {
		return err
	}
	compoundSplitter := compoundsplitting.NewSplitter(dict)

	vectorizer, err := NewVectorizer(s.rawContextionary, s.stopwordDetector, s.config, s.logger, NewSplitter(), extensionRetriever, compoundSplitter)
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

func (s *server) watchForSchemaChanges() {
	err := s.getInitialSchema(s.etcdClient)
	if err != nil {
		s.logger.WithField("action", "startup").
			WithError(err).Error("cannot retrieve initial schema")
		os.Exit(1)
	}

	rch := s.etcdClient.Watch(context.Background(), s.config.SchemaProviderKey)
	for wresp := range rch {
		for _, ev := range wresp.Events {
			s.unmarshalAndUpdateSchema(ev.Kv.Value)
		}
	}
}

type schemaState struct {
	ActionSchema *models.Schema `json:"action"`
	ThingSchema  *models.Schema `json:"thing"`
}

func (s *server) unmarshalSchema(bytes []byte) (*schema.Schema, error) {
	var state schemaState
	err := json.Unmarshal(bytes, &state)
	if err != nil {
		return nil, fmt.Errorf("could not parse the schema state: %s", err)
	}

	return &schema.Schema{
		Actions: state.ActionSchema,
		Things:  state.ThingSchema,
	}, nil
}

func (s *server) unmarshalAndUpdateSchema(bytes []byte) {
	schema, err := s.unmarshalSchema(bytes)
	if err != nil {
		s.logger.WithField("action", "schema-update").WithError(err).
			Error("could not unmarshal schema result")
	}

	s.schema = *schema
	err = s.buildContextionary()
	if err != nil {
		s.logger.WithField("action", "schema-update").WithError(err).
			Error("could not rebuild contextionary")
		return
	}

	s.logger.WithField("action", "schema-update").
		WithField("schema", s.schema).
		Info("succesfully updated schema")
}

func (s *server) getInitialSchema(c *clientv3.Client) error {
	res, err := c.Get(context.Background(), s.config.SchemaProviderKey)
	if err != nil {
		return fmt.Errorf("could not retrieve key '%s' from etcd: %v",
			s.config.SchemaProviderKey, err)
	}

	switch k := len(res.Kvs); {
	case k == 0:
		return nil
	case k == 1:
		s.unmarshalAndUpdateSchema(res.Kvs[0].Value)
		return nil
	default:
		return fmt.Errorf("unexpected number of results for key '%s', "+
			"expected to have 0 or 1, but got %d: %#v", s.config.SchemaProviderKey,
			len(res.Kvs), res.Kvs)
	}
}
