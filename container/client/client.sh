#!/bin/bash

ip route add 10.0.1.0/24 via 10.100.200.2

iperf -B 10.100.200.10 -c 10.0.1.10 -p 3000 --tos 0xB8 --trip-times -u -l 128 -b 128k > voip.txt &
iperf -B 10.100.200.10 -c 10.0.1.10 -p 2500 --trip-times --reverse -u -l 1450 -b 60M > hard.txt &