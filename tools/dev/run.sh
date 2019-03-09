export GO111MODULE=on
go run ./server 2>&1 | jq
