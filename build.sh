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

if [ -z "$FULL_VERSION_DOCKERFILE" ]
then
  FULL_VERSION_DOCKERFILE=Dockerfile.full
fi

echo "Build minimal version (english only)"
docker build -f Dockerfile.minimal --build-arg VERSION="$VERSION-minimal" -t "$DOCKER_ORG/$DOCKER_REPO:en$VERSION-minimal" .

echo "Build full versions"
echo "Build $LANGUAGE:"
full_version="${LANGUAGE}${VERSION}" 
docker buildx build --platform=linux/amd64,linux/arm64 \
  --load \
  -f "$FULL_VERSION_DOCKERFILE" \
  --build-arg VERSION="$full_version" \
  --build-arg MODEL_VERSION="$MODEL_VERSION" \
  --build-arg LANGUAGE="$LANGUAGE" \
  -t "$DOCKER_ORG/$DOCKER_REPO:$full_version" .
