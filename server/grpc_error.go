package main

import (
	"github.com/semi-technologies/contextionary/errors"
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
		// assume an unkown error is an internal error until we can be sure that
		// every component is returning correclty-typed errors
		// return status.Error(codes.Unknown, err.Error())
		return status.Error(codes.Internal, err.Error())
	}
}
