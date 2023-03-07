package main

import (
	"github.com/weaviate/contextionary/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GrpcErrFromTyped(err error) error {
	if err == nil {
		return nil
	}

	switch err.(type) {
	case errors.InvalidUserInput:
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Internal:
		return status.Error(codes.Internal, err.Error())
	case errors.NotFound:
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Unknown, err.Error())
	}
}
