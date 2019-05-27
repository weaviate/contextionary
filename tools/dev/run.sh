GO111MODULE=on \
  KNN_FILE="./tools/dev/example.knn" \
  IDX_FILE="./tools/dev/example.idx" \
  STOPWORDS_FILE="./tools/dev/stopwords.json" \
  go run ./server 2>&1 
