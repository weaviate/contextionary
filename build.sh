docker build -f Dockerfile.minimal -t "$DOCKER_ORG/$DOCKER_REPO:$VERSION-minimal" .
docker build -f Dockerfile.full -t "$DOCKER_ORG/$DOCKER_REPO:$VERSION" .
