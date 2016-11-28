FROM scratch

COPY build/ecs-gen-linux-amd64 /usr/bin/ecs-gen

WORKDIR /root/

CMD ["ecs-gen"]
