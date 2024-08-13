FROM golang:1.22 AS builder
LABEL maintainer="youssouf.drif"

RUN go install github.com/shynuu/trunks@latest

FROM debian:bookworm-slim
WORKDIR /trunks
RUN apt-get update && DEBIAN_FRONTEND=non-interactive apt-get install -qy --no-install-recommends --no-install-suggests \
    iproute2 iptables procps tcpdump iputils-ping \
    && rm -rf /var/lib/apt/lists/*
COPY --from=builder /go/bin/trunks /usr/local/sbin/trunks
ENTRYPOINT [ "trunks" ]
CMD [ "--help" ]
