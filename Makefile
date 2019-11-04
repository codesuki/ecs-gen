NAME = ecs-gen
VERSION = $(shell git describe --tags)

WORKDIR = /go/src/github.com/codesuki/ecs-gen

LDFLAGS = -X main.version=$(VERSION)

.PHONY: docker build clean deps

# docker:
	# docker build -t ecs-gen-builder:latest -f Dockerfile.build .
	# docker run --rm -v $(CURDIR):$(WORKDIR) ecs-gen-builder
	# docker build -t ecs-gen:latest -f Dockerfile .
	# docker run --rm ecs-gen:latest ecs-gen --version

build: deps
	for GOOS in darwin linux; do \
		for GOARCH in 386 amd64; do \
			GOOS=$$GOOS GOARCH=$$GOARCH go build -ldflags "$(LDFLAGS)" -v -o build/$(NAME)-$$GOOS-$$GOARCH ; \
		done \
	done

clean:
	rm -rf build

deps:
	glide install

zip:
	for file in build/*; do \
		zip -j -r "$${file}.zip" "$$file"; \
	done
