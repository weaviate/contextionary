#!/usr/bin/env bash
set -e

echo "Unpacking fixture vectors"
rm -f tools/dev/en_test-vectors-small.txt || true
bunzip2 -k tools/dev/en_test-vectors-small.txt.bz2

# Fake stopword removal by removing the first 10 words. This will become
# obsolete once we have released a new minimal c11y

# build stopword.json
cat tools/dev/en_test-vectors-small.txt | head | \
  while read -r word _; do echo "$word"; done | jq -nR '[inputs | select(length>0)] | { language: "en", words: . }'  > tools/dev/stopwords.json

# remove stop words
sed -i.bak 1,10d tools/dev/en_test-vectors-small.txt && rm tools/dev/en_test-vectors-small.txt.bak

if [ -f tools/dev/example.knn ]; then
  echo "Fixture contextionary already generated"
else
  go run contextionary/core/generator/cmd/generator.go \
    -c tools/dev/en_test-vectors-small.txt \
    -p tools/dev/example
fi
