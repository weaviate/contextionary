#!/bin/bash

set -e

# Jump to root directory
cd "$( dirname "${BASH_SOURCE[0]}" )"/..

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

docker tag "$DOCKER_ORG/$DOCKER_REPO:$VERSION-minimal" c11y-local-journeytest-minimal
docker tag "$DOCKER_ORG/$DOCKER_REPO:$VERSION" c11y-local-journeytest-full

echo "Cleaning up from previous runs"
docker-compose -f ./test/journey/docker-compose.yml down

echo "Starting containers"
docker-compose -f ./test/journey/docker-compose.yml up -d minimal full etcd

echo "Building tests"
docker-compose -f ./test/journey/docker-compose.yml build test-env 

echo "Running tests"
docker-compose -f ./test/journey/docker-compose.yml run test-env go test .

