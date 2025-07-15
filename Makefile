AWS_DEFAULT_REGION ?= eu-north-1
AWS_EXCHANGE_RATES_BUCKET ?= go-money-exchange-rates

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
	docker build -f ./build/Dockerfile.server --build-arg="SOURCE_PATH=cmd/server" -t ${DOCKER_SERVER_IMAGE_NAME} .

.PHONY: sam-deploy
sam-deploy:
	cd build && sam build && sam deploy --no-confirm-changeset --stack-name go-money-infra --capabilities CAPABILITY_IAM --resolve-s3 --parameter-overrides BucketName=${AWS_EXCHANGE_RATES_BUCKET} ExchangeRatesApiURL=${EXCHANGE_RATE_API_URL} --region ${AWS_DEFAULT_REGION}
