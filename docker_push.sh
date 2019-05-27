#!/bin/bash

echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
docker push "$DOCKER_ORG/$DOCKER_REPO:$TAG"
