FROM golang:1.7.3-alpine AS builder
RUN apk update && apk add --no-cache git make
WORKDIR /go/src/github.com/codesuki/ecs-gen
COPY glide.lock glide.yaml Makefile /go/src/github.com/codesuki/ecs-gen/
# to statically link
ENV CGO_ENABLED 0
RUN go get -u github.com/Masterminds/glide && make deps
COPY . .
RUN make build

FROM alpine:3.8
RUN apk update && apk add --no-cache ca-certificates openssl && update-ca-certificates
COPY --from=builder /go/src/github.com/codesuki/ecs-gen/build/ecs-gen-linux-amd64 /usr/bin/ecs-gen
CMD ["ecs-gen"]
