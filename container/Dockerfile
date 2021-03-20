FROM ubuntu:18.04
LABEL maintainer="youssouf.drif"

WORKDIR /trunks

COPY ./trunks ./trunks
RUN chmod +x trunks

RUN apt-get update
RUN apt-get install -y iproute2 iptables iputils-ping tcpdump iperf
