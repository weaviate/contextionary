#!/usr/bin/env bash
set -e

if [ -f tools/dev/en_test-vectors-small.txt ]; then
  echo "Already unpacked"
else
  echo "Unpacking fixture vectors"
  bunzip2 -k tools/dev/en_test-vectors-small.txt.bz2
fi

if [ -f tools/dev/example.knn ]; then
  echo "Fixture contextionary already generated"
else
  go run contextionary/core/generator/cmd/generator.go \
    -c tools/dev/en_test-vectors-small.txt \
    -p tools/dev/example
fi
