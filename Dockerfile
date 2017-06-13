FROM alpine

WORKDIR /root

RUN apk update && apk add ca-certificates openssl && update-ca-certificates

# download release of ecs-gen
ENV ECS_GEN_RELEASE 0.3.2
RUN wget https://github.com/codesuki/ecs-gen/releases/download/$ECS_GEN_RELEASE/ecs-gen-linux-amd64.zip && unzip ecs-gen-linux-amd64.zip && cp ecs-gen-linux-amd64 /usr/local/bin/ecs-gen

CMD ["ecs-gen"]
