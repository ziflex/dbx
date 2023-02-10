default: fmt lint test

test:
	go test ./...

lint:
	go vet ./... && \
	staticcheck -tests=false ./...

fmt:
	go fmt ./... && \
	goimports -w .