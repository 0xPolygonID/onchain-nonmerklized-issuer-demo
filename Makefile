.PHONY: run
run: build
	./onchain-non-merklized-issuer-demo

.PHONY: build
build:
	go build -o onchain-non-merklized-issuer-demo .

.PHONY: test
test:
	go test -v ./...

.PHONY: lint
lint:
	golangci-lint run
