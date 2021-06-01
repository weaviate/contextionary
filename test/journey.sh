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

if [ -z "$SOFTWARE_VERSION" ]
then
  SOFTWARE_VERSION=local
fi

if [ -z "$MODEL_VERSION" ]
then
  MODEL_VERSION=0.16.0
fi

if [ -z "$LANGUAGE" ]
then
  LANGUAGE=en
fi

VERSION="${MODEL_VERSION}-${SOFTWARE_VERSION}"

docker tag "$DOCKER_ORG/$DOCKER_REPO:${LANGUAGE}$VERSION-minimal" c11y-local-journeytest-minimal
docker tag "$DOCKER_ORG/$DOCKER_REPO:${LANGUAGE}$VERSION" c11y-local-journeytest-full

echo "Cleaning up from previous runs"
docker-compose -f ./test/journey/docker-compose.yml down

echo "Starting containers"
docker-compose -f ./test/journey/docker-compose.yml up -d minimal full weaviate

echo "Building tests"
docker-compose -f ./test/journey/docker-compose.yml build test-env 

echo "Running tests"
docker-compose -f ./test/journey/docker-compose.yml run test-env go test .

