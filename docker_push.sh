#!/bin/bash

set -e

echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
set -x
docker push "$DOCKER_ORG/$DOCKER_REPO:$VERSION-minimal"

