<p align="center">
<img width="500" alt="image" src="https://user-images.githubusercontent.com/41422704/111051142-37e67680-8451-11eb-9e1b-c3cdee7e0064.png">
</p>

<p align="center">
<a href="https://github.com/shynuu/trunks/releases"><img src="https://img.shields.io/badge/Release-v1.0-blue?logo=go" alt="Freecli 5G"/></a>
<img src="https://img.shields.io/badge/OS-Linux-g" alt="OS Linux"/>
<a href="https://github.com/shynuu/trunks/blob/master/LICENSE"><img src="https://img.shields.io/badge/license-MIT-lightgrey" alt="Apache 2 License"/></a>
</p>

- [Architecture](#architecture)
- [Requirements](#requirements)
- [Installation and usage](#installation-and-usage)
- [Features](#features)
  - [Bandwidth](#bandwidth)
  - [Delay](#delay)
  - [ACM](#acm)

## Architecture

Trunks simulates a simple DVB-S2/RCS2 satellite system using native linux tools tc and iptables, the following figures depicts the architecture of the software.

## Requirements

**Hardware:** Linux host with two network interfaces UP and configured

**Software:** `tc`, `iptables`, `go`

Tested using `iperf` between 2 hosts with the following testbed:

- Trunks running under Ubuntu 18.04 VirtualBox VM 1 CPU 1 RAM, golang 1.16.2
- Host 1 running under Ubuntu 18.04 VirtualBox VM 1 CPU 1 RAM iperf server
- Host 2 running under Ubuntu 18.04 VirtualBox VM 1 CPU 1 RAM iperf client

## Installation and usage

Steps for installation:

```bash
git clone 
```


## Features

```yaml
# set the network device for satellite terminal and gateway.
nic:
  st: enp0s8
  gw: enp0s9

# configure the forward and return links available bandwidth in Mbits/s
bandwidth:
  forward: 80
  return: 20

# configure the delay according to the GEO, MEO or LEO satellite and the offset, real delay = delay + or - offset
delay:
  value: 20
  offset: 10

# set the ACM simulation values
acm:
  - weight: 1
    duration: 10
  - weight: 0.8
    duration: 10
  - weight: 0.9
    duration: 10
  - weight: 0.5
    duration: 10
  - weight: 0.7
    duration: 10
```

### Bandwidth

### Delay

### ACM