package main

import (
	"fmt"
	"net"
	"os"

	pb "github.com/semi-technologies/contextionary/contextionary"
	core "github.com/semi-technologies/contextionary/contextionary/core"
	"github.com/semi-technologies/contextionary/server/config"
	"github.com/semi-technologies/weaviate/entities/schema"
	"github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"
)

func main() {
	server := new()
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

	config *config.Config

	logger logrus.FieldLogger
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
