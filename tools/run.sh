#!/bin/sh

# Starts contextionary with existing data files

# If you need to download the data files, run tools/native_build_contextionary.sh

#Set environment variables
export KNN_FILE=./data/contextionary.knn
export IDX_FILE=./data/contextionary.idx
export STOPWORDS_FILE=./data/stopwords.json
export COMPOUND_SPLITTING_DICTIONARY_FILE=./data/splitter_dict.csv

#Run the server
./contextionary-server

