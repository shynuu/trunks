#/bin/bash

# ======= QoS Model used when running Trunks with the `--qos` option =======

ST_IFACE=enp0s8
GW_IFACE=enp0s9

UP=20mbit
UP_VOIP=2mbit
UP_REST=18mbit
RET=50mbit
RET_VOIP=2mbit
RET_REST=48mbit
DELAY=20ms
OFFSET=2ms

# configure redirection for tc
sudo iptables -F -t mangle
sudo iptables -t mangle -A PREROUTING -i $ST_IFACE -j MARK --set-mark 10
sudo iptables -t mangle -A PREROUTING -i $ST_IFACE -m dscp --dscp 0x2c -j MARK --set-mark 11
sudo iptables -t mangle -A PREROUTING -i $ST_IFACE -m dscp --dscp 0x2e -j MARK --set-mark 11

sudo iptables -t mangle -A PREROUTING -i $GW_IFACE -j MARK --set-mark 20
sudo iptables -t mangle -A PREROUTING -i $GW_IFACE -m dscp --dscp 0x2c -j MARK --set-mark 21
sudo iptables -t mangle -A PREROUTING -i $GW_IFACE -m dscp --dscp 0x2e -j MARK --set-mark 21

sudo iptables -L -t mangle -nv

# =========== Configure rules for return link ============
# configure TC
sudo tc qdisc del dev $GW_IFACE root
sudo tc filter del dev $GW_IFACE

# configure rules for TC
sudo tc qdisc add dev $GW_IFACE root handle 1:0 htb default 20

sudo tc class add dev $GW_IFACE parent 1:0 classid 1:1 htb rate $UP

sudo tc class add dev $GW_IFACE parent 1:1 classid 1:10 htb rate $UP_VOIP prio 0
sudo tc qdisc add dev $GW_IFACE parent 1:10 handle 110: netem delay $DELAY $OFFSET distribution normal


sudo tc class add dev $GW_IFACE parent 1:1 classid 1:20 htb rate $UP_REST prio 1
sudo tc qdisc add dev $GW_IFACE parent 1:20 handle 120: netem delay $DELAY $OFFSET distribution normal

sudo tc filter add dev $GW_IFACE protocol ip parent 1:0 prio 0 handle 11 fw flowid 1:10
sudo tc filter add dev $GW_IFACE protocol ip parent 1:0 prio 1 handle 10 fw flowid 1:20

sudo tc -s qdisc ls dev $GW_IFACE
sudo tc -s class ls dev $GW_IFACE
sudo tc -s filter ls dev $GW_IFACE

# =========== Configure rules for forward link ============
# configure TC
sudo tc qdisc del dev $ST_IFACE root
sudo tc filter del dev $ST_IFACE

# configure rules for TC
sudo tc qdisc add dev $ST_IFACE root handle 1:0 htb default 20

sudo tc class add dev $ST_IFACE parent 1:0 classid 1:1 htb rate $RET

sudo tc class add dev $ST_IFACE parent 1:1 classid 1:10 htb rate $RET_VOIP prio 0
sudo tc qdisc add dev $ST_IFACE parent 1:10 handle 110: netem delay $DELAY $OFFSET distribution normal

sudo tc class add dev $ST_IFACE parent 1:1 classid 1:20 htb rate $RET_REST prio 1
sudo tc qdisc add dev $ST_IFACE parent 1:20 handle 120: netem delay $DELAY $OFFSET distribution normal

sudo tc filter add dev $ST_IFACE protocol ip parent 1:0 prio 0 handle 21 fw flowid 1:10
sudo tc filter add dev $ST_IFACE protocol ip parent 1:0 prio 1 handle 20 fw flowid 1:20

sudo tc -s qdisc ls dev $ST_IFACE
sudo tc -s class ls dev $ST_IFACE
sudo tc -s filter ls dev $ST_IFACE
