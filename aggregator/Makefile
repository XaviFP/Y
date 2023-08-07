aggregator:
	go build -o bin/aggregator ./cmd/aggregator/ ;

test:
	go test ./... -p=1 -coverprofile=coverage.out *.go

coverage: test
	go tool cover -html=coverage.out

