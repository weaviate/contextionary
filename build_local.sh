#!/bin/bash

set -e

export LANGUAGES="en"
export VERSION="0.16.0-v0.4.17"
export MODEL_VERSION="0.16.0"

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

if [ -z "$FULL_VERSION_DOCKERFILE" ]
then
  FULL_VERSION_DOCKERFILE=Dockerfile.full
fi

echo "Build full versions"
for lang in $LANGUAGES; do
  echo "Build $lang:"
  full_version="${lang}${VERSION}" 
  docker build -f "$FULL_VERSION_DOCKERFILE" \
    --build-arg VERSION="$full_version" \
    --build-arg MODEL_VERSION="$MODEL_VERSION" \
    --build-arg LANGUAGE="$lang" \
    -t "$DOCKER_ORG/$DOCKER_REPO:$full_version" .
done
