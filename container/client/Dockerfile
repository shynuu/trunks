FROM debian:buster-slim
LABEL maintainer="youssouf.drif"

RUN apt-get update && apt-get install -y build-essential git iproute2 tcpdump iptables iputils-ping
WORKDIR /iperf
RUN git clone https://git.code.sf.net/p/iperf2/code iperf2-source
WORKDIR /iperf/iperf2-source
RUN ./configure
RUN make
RUN make install

WORKDIR /app
COPY ./client.sh client.sh
COPY ./server.sh server.sh