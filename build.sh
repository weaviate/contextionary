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

echo "Build minimal version (english only)"
docker build -f Dockerfile.minimal --build-arg VERSION="$VERSION-minimal" -t "$DOCKER_ORG/$DOCKER_REPO:$VERSION-minimal" .

echo "Build full versions"
for lang in $LANGUAGES; do
  echo "Build $lang:"
  full_version="${lang}${VERSION}" 
  docker build -f Dockerfile.full \
    --build-arg VERSION="$full_version" \
    --build-arg LANGUAGE="$lang" \
    -t "$DOCKER_ORG/$DOCKER_REPO:$full_version" .
done
