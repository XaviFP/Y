publisher:
	go build -o bin/publisher ./cmd/publisher/ ;

test:
	go test ./... -p=1 -coverprofile=coverage.out *.go

coverage: test
	go tool cover -html=coverage.out

