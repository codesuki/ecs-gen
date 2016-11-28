FROM scratch

COPY build/ecs-gen /usr/bin/

WORKDIR /root/

CMD ["ecs-gen"]
