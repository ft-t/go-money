AWS_DEFAULT_REGION ?= eu-north-1
AWS_EXCHANGE_RATES_BUCKET ?= go-money-exchange-rates
DOCKER_SERVER_IMAGE_NAME ?= "go-money-server:latest"

VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT_SHA ?= $(shell git rev-parse --short HEAD)

#

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


.PHONY: build-docker
build-docker:
	docker build -f ./build/Dockerfile.server --build-arg="SOURCE_PATH=cmd/server" --build-arg="VERSION=${VERSION}" --build-arg="COMMIT_SHA=${COMMIT_SHA}" -t ${DOCKER_SERVER_IMAGE_NAME} .

.PHONY: sam-deploy
sam-deploy:
	cd build && sam build && sam deploy --no-confirm-changeset --stack-name go-money-infra --capabilities CAPABILITY_IAM --resolve-s3 --parameter-overrides BucketName=${AWS_EXCHANGE_RATES_BUCKET} ExchangeRatesApiURL=${EXCHANGE_RATE_API_URL} --region ${AWS_DEFAULT_REGION}
