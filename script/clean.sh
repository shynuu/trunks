#/bin/bash

# ======= Clean script using bash =======

ST_IFACE=enp0s8
GW_IFACE=enp0s9

# configure redirection for tc
sudo iptables -F -t mangle
sudo tc qdisc del dev $GW_IFACE root
sudo tc filter del dev $GW_IFACE

sudo tc qdisc del dev $ST_IFACE root
sudo tc filter del dev $ST_IFACE

sudo tc -s qdisc ls dev $GW_IFACE
sudo tc -s class ls dev $GW_IFACE
sudo tc -s filter ls dev $GW_IFACE

sudo tc -s qdisc ls dev $ST_IFACE
sudo tc -s class ls dev $ST_IFACE
sudo tc -s filter ls dev $ST_IFACE
