#!/bin/bash

ip route add 10.100.200.0/24 via 10.0.1.2

iperf -s -B 10.0.1.10 -p 3000 --tos 0xB8 & 
iperf -s -B 10.0.1.10 -p 2500 &
