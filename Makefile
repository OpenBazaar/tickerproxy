DOCKER_PROFILE ?= openbazaar
DOCKER_VERSION ?= $(shell git describe --tags --abbrev=0)
DOCKER_IMAGE_NAME ?= $(DOCKER_PROFILE)/tickerproxy:$(DOCKER_VERSION)

docker:
	docker build -t $(DOCKER_IMAGE_NAME) .

push_docker:
	docker push $(DOCKER_IMAGE_NAME)