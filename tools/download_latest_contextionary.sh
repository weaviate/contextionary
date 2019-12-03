#!/bin/bash

set -e

VECTORDB_VERSION=0.8.0
LANGUAGE=${1}

rm -rf ./data && mkdir ./data

# Download the latest files and remove old ones
for FILE in stopwords.json contextionary.idx contextionary.knn; do
    echo "Start Downloading $FILE" && \
    wget --quiet -O ./data/$FILE https://c11y.semi.technology/$VECTORDB_VERSION/$LANGUAGE/$FILE && \
    echo "$FILE = done" &
done 

# Wait to finish download
wait

echo "Done downloading open source contextionary v$VECTORDB_VERSION."
exit 0
