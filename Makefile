.DEFAULT_GOAL := help

DOCKER_PROFILE ?= openbazaar
DOCKER_VERSION ?= $(shell git describe --tags --abbrev=0)
DOCKER_IMAGE_NAME ?= $(DOCKER_PROFILE)/ticker:$(DOCKER_VERSION)

LAMBDA_FILENAME ?= update_price_ticker.zip
LAMBDA_PATH ?= lambdas
LAMBDA_DEPLOY_BUCKET ?= deploy-bucket

.PHONY: lambda
lambda: ## Build lambda package
	mkdir -p dist/lambda
	go build -o dist/lambda/main ./lambda
	cd dist/lambda && zip -r $(LAMBDA_FILENAME) main

deploy_lambda: ## Deploy lambda artifact
	aws s3api put-object --bucket $(LAMBDA_DEPLOY_BUCKET) --key $(LAMBDA_PATH)/$(LAMBDA_FILENAME) --body dist/lambda/$(LAMBDA_FILENAME)

binary: ## Build fetch binary
	go build -o dist/fetch ./fetch

docker: ## Build docker image
	docker build -t $(DOCKER_IMAGE_NAME) .

push_docker: ## Push docker image to remote repository
	docker push $(DOCKER_IMAGE_NAME)

tests: ## Run tests
	go test -v -cover .

profile_tests: ## Run tests with coverage profiling
	go test -v -coverprofile=coverage.out .
	go tool cover -html=coverage.out

clean:  ## Remove all built artifacts
	rm -r ./dist
	docker rmi $(DOCKER_IMAGE_NAME)

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
