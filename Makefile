services:
	$(MAKE) -C aggregator
	$(MAKE) -C publisher

test:
	go test ./... -p=1 -coverprofile=coverage.out *.go

coverage: test
	go tool cover -html=coverage.out

run: services
	docker compose up --build

