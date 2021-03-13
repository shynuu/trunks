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

Trunks is a lightweight DVB-S2/RCS2 satellite system simulator using native linux tools tc and iptables, the following figures depicts the architecture of the software.

It can run under a single VM or Docker.

## Requirements

**Hardware:** Linux host with two network interfaces UP and configured

**Software:** `tc`, `iptables`, `go`

Tested using `iperf` between 2 hosts with the following testbed:

- Trunks running under Ubuntu 18.04 VirtualBox VM 1 CPU 1 RAM, golang 1.16.2
- Host 1 running under Ubuntu 18.04 VirtualBox VM 1 CPU 1 RAM iperf server/client
- Host 2 running under Ubuntu 18.04 VirtualBox VM 1 CPU 1 RAM iperf client/server

## Installation and usage

Steps for installation:

```bash
git clone https://github.com/shynuu/trunks
go build -o trunk -x main.go
```

Launch trunks with sudo privilege (required for enabling the forwarding and interacting with iptables/tc):

```bash
sudo ./trunk --config config/trunks.yaml --flush --acm
```

Launch options are detailed below:

```bash
NAME:
   trunks - a simple DVB-S2/DVB-RCS2 simulator

USAGE:
   trunk [global options] command [command options] [arguments...]

AUTHOR:
   Youssouf Drif

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config FILE  Load configuration from FILE (default: not set)
   --flush        Flush IPTABLES table mangle and clear all TC rules (default: false)
   --acm          Activate the ACM simulation (default: not activated)
   --help, -h     show help (default: false)

```

## Features

You need to associate each network interface with a satellite component. The component/interface association is important as it will determine the forward and return link process. Trunks code is based on this [script](script/static_simulation.sh).

Change the config `config/trunks.yaml` file or create a new one.

```yaml
nic:
  st: enp0s8
  gw: enp0s9
```

### Bandwidth

You can set the bandwidth (in Mbit/s) for the forward and return link

```yaml
bandwidth:
  forward: 80
  return: 20
```

### Delay

You can set the delay (in ms) to simulate a LEO, MEO or GEO altitude. The delay changes during the simulation and is comprised between 

`delay - offset <= value <= delay + offset`

```yaml
delay:
  value: 20
  offset: 10
```

### ACM

When you launch Trunks with the option `--acm`, the ACM mechanism of DVB-S2/RCS2 systems is simulated.

The maximum bandwidth of the forward and return link will vary in function of values set in the config file:

```yaml
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

The program picks a random tuple of this list and weights the maximum bandwidth (`forward = weight * forward` and `return = weight * return`) to the link for the `duration` specified (in second). At the end of this duration, it randomly picks another tuple and restarts the process.