#!/bin/bash

set -e

language=${1}

# Get dictionaries
git clone https://github.com/LibreOffice/dictionaries.git

index_file="/app/data/contextionary.idx"
aff_file=""
dic_file=""

if [ "$language" == "en" ]; then
  aff_file="/app/dictionaries/en/en_US.aff"
  dic_file="/app/dictionaries/en/en_US.aff"
fi
if [ "$language" == "de" ]; then
  aff_file="/app/dictionaries/de/de_DE_frami.aff"
  dic_file="/app/dictionaries/de/de_DE_frami.dic"
fi
if [ "$language" == "nl" ]; then
  aff_file="/app/dictionaries/nl_NL/nl_NL.aff"
  dic_file="/app/dictionaries/nl_NL/nl_NL.dic"
fi

if [ "$aff_file" == "" ]; then
  echo "Missing dictionary for preprocessor see process_splitter_dict.sh"
  exit 3
fi

go run main/splitter_preprocessor.go "$index_file" "$dic_file" "$aff_file" "/app/data/splitter_dict.csv"

