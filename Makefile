.PHONY: lint
lint:
	golangci-lint run

.PHONY: lint-dev
lint-dev:
	golangci-lint run --tests=false

.PHONY: generate
generate:
	go generate ./...

.PHONY: update
update: update-pb

.PHONY: update-pb
update-pb:
	go get buf.build/gen/go/xskydev/go-money-pb/connectrpc/go@master
	go mod tidy

.PHONY: test
test:
	AUTO_CREATE_CI_DB=true go test ./...


DOCKER_SERVER_IMAGE_NAME ?= "go-money-server:latest"

.PHONY: build-docker
build-docker:
	docker build -f ./cmd/server/Dockerfile --build-arg="SOURCE_PATH=cmd/server" -t ${DOCKER_SERVER_IMAGE_NAME} .
