#!/bin/bash

set -e

language=${1}
index_file=${2}

# Get dictionaries
git clone https://github.com/LibreOffice/dictionaries.git

aff_file=""
dic_file=""

if [ "$language" == "en" ]; then
  aff_file="/app/dictionaries/en/en_US.aff"
  dic_file="/app/dictionaries/en/en_US.dic"
fi
if [ "$language" == "de" ]; then
  aff_file="/app/dictionaries/de/de_DE_frami.aff"
  dic_file="/app/dictionaries/de/de_DE_frami.dic"
fi
if [ "$language" == "nl" ]; then
  aff_file="/app/dictionaries/nl_NL/nl_NL.aff"
  dic_file="/app/dictionaries/nl_NL/nl_NL.dic"
fi
if [ "$language" == "it" ]; then
  aff_file="/app/dictionaries/it_IT/it_IT.aff"
  dic_file="/app/dictionaries/it_IT/it_IT.dic"
fi
if [ "$language" == "cs" ]; then
  aff_file="/app/dictionaries/cs_CZ/cs_CZ.aff"
  dic_file="/app/dictionaries/cs_CZ/cs_CZ.dic"
fi

if [ "$aff_file" == "" ]; then
  echo "Missing dictionary for preprocessor see process_splitter_dict.sh"
  exit 3
fi

echo "Building dict with:"
go run main/splitter_preprocessor.go "$index_file" "$dic_file" "$aff_file" "/app/data/splitter_dict.csv"

