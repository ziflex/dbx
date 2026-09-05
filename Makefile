default: fmt-check lint test

test:
	go test ./...

lint:
	golangci-lint run

fmt:
	go fmt ./... && \
	goimports -w -local github.com/ziflex/dbx .

fmt-check:
	@files="$$(gofmt -l .)"; if [ -n "$$files" ]; then echo "gofmt required for:"; echo "$$files"; exit 1; fi
	@files="$$(goimports -l -local github.com/ziflex/dbx .)"; if [ -n "$$files" ]; then echo "goimports required for:"; echo "$$files"; exit 1; fi

test-race:
	go test -race ./...
