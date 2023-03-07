package main

import (
	"fmt"
	"net"
	"os"

	"github.com/sirupsen/logrus"
	pb "github.com/weaviate/contextionary/contextionary"
	core "github.com/weaviate/contextionary/contextionary/core"
	"github.com/weaviate/contextionary/extensions"
	"github.com/weaviate/contextionary/server/config"
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
	// to be used to serve rpc requests, combination of the raw contextionary
	// and the schema
	combinedContextionary core.Contextionary

	// initialized at startup, to be used to build the
	// schema contextionary
	rawContextionary core.Contextionary

	config *config.Config

	logger logrus.FieldLogger

	// ucs
	extensionStorer      *extensions.Storer
	extensionLookerUpper extensionLookerUpper
	stopwordDetector     stopwordDetector
	vectorizer           *Vectorizer
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

	err = s.init()
	if err != nil {
		logger.
			WithError(err).
			Errorf("cannot start up")
		os.Exit(1)
	}

	return s
}
