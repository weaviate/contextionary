package main

import (
	"fmt"
	"net"
	"os"

	"github.com/coreos/etcd/clientv3"
	"github.com/semi-technologies/contextionary/adapters/repos"
	pb "github.com/semi-technologies/contextionary/contextionary"
	core "github.com/semi-technologies/contextionary/contextionary/core"
	"github.com/semi-technologies/contextionary/server/config"
	"github.com/semi-technologies/weaviate/entities/schema"
	"github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"
)

// Version is filled through a build arg
var Version string

func main() {
	server := new()
	server.logger.WithField("version", Version).Info()
	grpcServer := grpc.NewServer()
	pb.RegisterContextionaryServer(grpcServer, server)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", server.config.ServerPort))
	if err != nil {
		server.logger.Errorf("can't listen on port: %s", err)
		os.Exit(1)
	}

	grpcServer.Serve(lis)
}

type server struct {
	schema schema.Schema

	// to be used to serve rpc requests, combination of the raw contextionary
	// and the schema
	combinedContextionary core.Contextionary

	// initialized at startup, to be used to build the
	// schema contextionary
	rawContextionary core.Contextionary

	stopwordDetector stopwordDetector

	config *config.Config

	logger logrus.FieldLogger

	etcdClient *clientv3.Client
}

// new gRPC server to serve the contextionary
func new() *server {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	cfg, err := config.New(logger)
	if err != nil {
		logger.
			WithError(err).
			Errorf("cannot start up")
		os.Exit(1)
	}

	loglevel, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		logger.
			WithError(err).
			Errorf("cannot start up")
		os.Exit(1)
	}
	logger.SetLevel(loglevel)
	logger.WithField("log_level", loglevel.String()).Info()

	s := &server{
		config: cfg,
		logger: logger,
	}

	etcdClient, err := clientv3.New(clientv3.Config{Endpoints: []string{s.config.SchemaProviderURL}})
	if err != nil {
		s.logger.WithField("action", "startup").
			WithError(err).Error("cannot construct etcd client")
		os.Exit(1)
	}

	s.etcdClient = etcdClient

	err = s.init()
	if err != nil {
		logger.
			WithError(err).
			Errorf("cannot start up")
		os.Exit(1)
	}

	er := repos.NewEtcdExtensionRepo(etcdClient, logger, cfg)
	_ = er // TODO: pass to extension looker upper

	return s
}
