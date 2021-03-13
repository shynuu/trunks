#/bin/bash

ST_IFACE=enp0s8
GW_IFACE=enp0s9

UP=20mbit
RET=50mbit
DELAY=250ms
OFFSET=100ms

# configure redirection for tc
sudo iptables -F -t mangle
sudo iptables -t mangle -A PREROUTING -i $ST_IFACE -j MARK --set-mark 10
sudo iptables -t mangle -A PREROUTING -i $GW_IFACE -j MARK --set-mark 20
sudo iptables -L -t mangle -nv

# =========== Configure rules for return link ============
# configure TC
sudo tc qdisc del dev $GW_IFACE root
sudo tc filter del dev $GW_IFACE

# configure rules for TC
sudo tc qdisc add dev $GW_IFACE root handle 1:0 htb default 30
sudo tc class add dev $GW_IFACE parent 1:0 classid 1:1 htb rate $UP
sudo tc qdisc add dev $GW_IFACE parent 1:1 handle 2:0 netem delay $DELAY $OFFSET distribution normal
sudo tc filter add dev $GW_IFACE protocol ip parent 1:0 prio 1 handle 10 fw flowid 1:1
sudo tc -s qdisc ls dev $GW_IFACE
sudo tc -s class ls dev $GW_IFACE
sudo tc -s filter ls dev $GW_IFACE

# =========== Configure rules for forward link ============
# configure TC
sudo tc qdisc del dev $ST_IFACE root
sudo tc filter del dev $ST_IFACE

# configure rules for TC
sudo tc qdisc add dev $ST_IFACE root handle 1:0 htb default 30
sudo tc class add dev $ST_IFACE parent 1:0 classid 1:1 htb rate $RET
sudo tc qdisc add dev $ST_IFACE parent 1:1 handle 2:0 netem delay $DELAY $OFFSET distribution normal
sudo tc filter add dev $ST_IFACE protocol ip parent 1:0 prio 1 handle 20 fw flowid 1:1
sudo tc -s qdisc ls dev $ST_IFACE
sudo tc -s class ls dev $ST_IFACE
sudo tc -s filter ls dev $ST_IFACE