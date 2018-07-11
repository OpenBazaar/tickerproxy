.DEFAULT_GOAL := help

DOCKER_PROFILE ?= openbazaar
DOCKER_VERSION ?= $(shell git describe --tags --abbrev=0)
DOCKER_IMAGE_NAME ?= $(DOCKER_PROFILE)/ticker:$(DOCKER_VERSION)

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
