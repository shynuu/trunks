<p align="center">
<img width="500" alt="image" src="https://user-images.githubusercontent.com/41422704/111051142-37e67680-8451-11eb-9e1b-c3cdee7e0064.png">
</p>

<p align="center">
<a href="https://github.com/shynuu/trunks/releases/tag/v2.0"><img src="https://img.shields.io/badge/Release-v2.0-blue?logo=go" alt="Freecli 5G"/></a>
<img src="https://img.shields.io/badge/OS-Linux-g" alt="OS Linux"/>
<a href="https://github.com/shynuu/trunks/blob/master/LICENSE"><img src="https://img.shields.io/badge/license-MIT-lightgrey" alt="MIT License"/></a>
</p>

- [Architecture](#architecture)
- [Requirements](#requirements)
- [Installation and usage](#installation-and-usage)
- [Features](#features)
  - [Bandwidth](#bandwidth)
  - [Delay](#delay)
  - [ACM](#acm)
  - [Quality of Service (QoS)](#quality-of-service-qos)
- [Docker](#docker)

## Architecture

Trunks is a lightweight DVB-S2/RCS2 satellite system simulator using the native linux tools tc and iptables. The following figure depicts the architecture of the software.

It can run under a single VM or Docker.

<img width="800" alt="image" src="https://user-images.githubusercontent.com/41422704/111052860-3fad1780-845f-11eb-9e6b-c24d55909ee1.png">
</p>


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
sudo ./trunk --config config/trunks.yaml --flush --acm --qos
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
   --config FILE                   Load configuration from FILE (default: not set)
   --logs value                    Log path for the log file (default: not set)
   --flush                         Flush IPTABLES table mangle and clear all TC rules (default: false)
   --acm                           Activate the ACM simulation (default: not activated)
   --qos                           Process traffic using QoS (default: not activated)
   --disable-kernel-version-check  Disable check for bugged kernel versions (default: kernel version check enabled)
   --help, -h                      show help (default: false)

```

## Features

Configuration: change the config `config/trunks.yaml` file or create a new one.

You need to associate each network interface with a satellite component. The component/interface association is important as it will determine the forward and return link process. Trunks code is based on this [script](script/static_simulation.sh). You can either set the L2 interface name or the IP address of the interface. Both interfaces must already be configured and UP.

```yaml
nic:
  st: enp0s8
  gw: enp0s9
```

### Bandwidth

You can set the bandwidth (in Mbit/s) for the forward and return link. Minimum required is 2 Mbit/s for the forward and return link.

```yaml
bandwidth:
  forward: 80
  return: 20
```

### Delay

You can set the delay (in ms) to simulate a LEO, MEO or GEO altitude. The delay changes during the simulation and is comprised between `delay - offset <= value <= delay + offset`.

```yaml
delay:
  value: 20
  offset: 10
```

### ACM

When you launch Trunks with the option `--acm`, the ACM mechanism of DVB-S2/RCS2 systems is simulated.

The maximum bandwidth of the forward and return link will vary according to the values set in the config file:

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

The program picks a random tuple of this list and weights the maximum bandwidth (`forward = weight * forward` and `return = weight * return`) of the link for the specified `duration` (in seconds). At the end of this duration, it randomly picks another tuple and restarts the process.

### Quality of Service (QoS)

You can activate the QoS with the option `--qos`. Basically the traffic coming in and out of trunks with the PHB EF (DSCP 0x2e) and PHB VA (DSCP 0x2c) will be processed in a dedicated HTB. This HTB has the highest priority value and has a 1 Mbit/s bandwidth reserved.


## Docker

You can build the docker image by running this [script](container/build.sh)

Then try the [docker-compose](container/docker-compose.yaml) example executing the following commands:

```bash
cd container
docker-compose up -d
```

You still must configure the routes inside the client and server containers.

## Known issues
Some versions of the kernel are known to crash when offset delay is enabled. As a protection, offset delay is by default disabled if you are running on affected versions. You can bypass protection by using the flag `--disable-kernel-version-check`, but this is NOT recommended. See [this issue](https://github.com/shynuu/trunks/issues/6) for details.
