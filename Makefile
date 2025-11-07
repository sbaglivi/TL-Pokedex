IMAGE_NAME := tl-pokedex
PLATFORM := --platform linux/amd64
PORT := 3000
REPO_ROOT := $(shell git rev-parse --show-toplevel)
BIN_OUT := bin/pokedex

docker-build:
	docker build -t ${IMAGE_NAME} ${PLATFORM} ${REPO_ROOT}

docker-run: docker-build
	docker run -it --rm ${PLATFORM} -p ${PORT}:3000 ${IMAGE_NAME}

run:
	PORT=${PORT} go run ${REPO_ROOT}/main.go

build:
	go build -o ${REPO_ROOT}/${BIN_OUT} ${REPO_ROOT}/main.go

test:
	go test ${REPO_ROOT}/...
	go test -race ${REPO_ROOT}/cache