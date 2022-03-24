lint:
	golangci-lint run

bench:
	$(ENVS) go test -run=NONE -bench=. -benchmem ./...

test:
	$(ENVS) go test -race -cover ./...

.PHONY: lint bench test
