FROM golang:1.20-bullseye as builder

COPY . /build/
RUN  cd /build/ && go build -ldflags "-w -s" -o ./_app .


FROM debian:bullseye-slim
USER root

RUN apt-get update && apt-get install -y \
    ca-certificates curl openssh-client &&\
    apt-get autoremove && apt-get clean && \
    rm -rf /tmp/* /var/tmp/* /var/lib/apt/lists/*

COPY --from=builder /build/_app  /app/app
COPY --from=builder /build/static  /app/static
COPY --from=builder /build/doc/config.toml  /app/config.toml

COPY start.sh       /app/start.sh

WORKDIR /app
CMD ["./start.sh"]

