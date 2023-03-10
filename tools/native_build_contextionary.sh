#!/bin/sh

#Download contextionary
LANGUAGE=en
MODEL_VERSION=0.16.0
./tools/download_contextionary.sh "$LANGUAGE" "$MODEL_VERSION"

#Build the server
VERSION=1.2.0
CGO_ENABLED=1 go build -o ./contextionary-server -a -tags netgo -ldflags "-w -X main.Version=$VERSION" ./server

#Generate contextionary
tools/dev/gen_simple_contextionary.sh

#Preprocess splitter dictionary
/bin/bash ./tools/preprocess_splitter_dict.sh "$LANGUAGE" "./data/contextionary.idx"

#Copy files to Alpine image
cp ./contextionary-server $PWD

#Set environment variables
export KNN_FILE=./data/contextionary.knn
export IDX_FILE=./data/contextionary.idx
export STOPWORDS_FILE=./data/stopwords.json
export COMPOUND_SPLITTING_DICTIONARY_FILE=./data/splitter_dict.csv

#Run the server
./contextionary-server
