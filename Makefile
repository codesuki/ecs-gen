NAME = ecs-gen
VERSION = $(shell git describe --tags)

WORKDIR = /go/src/github.com/codesuki/ecs-gen

LDFLAGS = -X main.version=$(VERSION)

.PHONY: docker build clean deps

docker:
	docker build -t ecs-gen-builder:latest -f Dockerfile.build .
	docker run --rm -v $(CURDIR):$(WORKDIR) ecs-gen-builder
	docker build -t ecs-gen:latest -f Dockerfile .
	docker run --rm ecs-gen:latest ecs-gen --version

build: deps
	go build -ldflags "$(LDFLAGS)" -o build/$(NAME)

clean:
	rm -rf build

deps:
	glide install
