GO111MODULE=on \
  KNN_FILE="./tools/dev/example.knn" \
  IDX_FILE="./tools/dev/example.idx" \
  STOPWORDS_FILE="./tools/dev/stopwords.json" \
  SCHEMA_PROVIDER_URL="localhost:2379" \
  go run ./server 2>&1 
