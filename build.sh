#!/bin/bash

set -e

# set some defaults so we can also run locally
if [ -z "$DOCKER_ORG" ]
then
  DOCKER_ORG=semitechnologies
fi

if [ -z "$DOCKER_REPO" ]
then
  DOCKER_REPO=contextionary
fi

if [ -z "$VERSION" ]
then
  VERSION=local
fi

docker build -f Dockerfile.minimal -t "$DOCKER_ORG/$DOCKER_REPO:$VERSION-minimal" .
docker build -f Dockerfile.full -t "$DOCKER_ORG/$DOCKER_REPO:$VERSION" .
