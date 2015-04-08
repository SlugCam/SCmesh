# SCmesh

**Warning, this code is not suitable for anything at the moment. It is under heavy
development.**

[![Build Status](https://travis-ci.org/SlugCam/SCmesh.svg?branch=master)](https://travis-ci.org/SlugCam/SCmesh)
[![Coverage Status](https://coveralls.io/repos/SlugCam/SCmesh/badge.svg)](https://coveralls.io/r/SlugCam/SCmesh)

SCmesh is a daemon that implements mesh networking over a serial based wireless module (such as the WiFly). It is used to support an ad hoc networking mode in the SlugCam system.

## WiFly Configuration

These steps should get the WiFly (in particular the RN-171) ready to run this software. Note that a tool to send commands to multiple serial interfaces can be useful. One trick that can be used is tmux with connections open in all the panes of a window. `:setw synchronize-panes` will then send input to all panes.

First connect to the WiFly using:

```sh
sudo minicom -D /dev/ttyAMA0 -b 9600
```


Now we should load the firmware version that allows ad hoc networking. First connect to a network with internet access. For example for WPA networks when WiFly is in default state and using DHCP:

```
set wlan ssid <ssid>
set wlan phrase <phrase>
join
```

*Note:* If the WiFly is not in the default state the dhcp setting, channel, and join may need to be set. Join 1 joins saved config on reboot. Chan 0 is automatic.

Now we need to update the firmware. These commands are from the WiFly command reference, except we use the archive directory to retrieve an older version of the firmware with ad hoc capability.

```
set dns backup rn.microchip.com
set ftp addr 0
set ftp username roving
set ftp pass Pass123
set ftp dir <dir>
ftp update <filename>
factory RESET
```

Some of these values may already be set. `dir` would normally be `./public`, but we will use `./archive`. Version 2.36 is the latest version with ad hoc capability. File for RN-171 is `wifly7-236.img`.

If the update procedure leads to a version downgrade the configuration file may not be compatible. `del config` followed by a `reboot` will start fresh.

Next we change the baud rate. Enter command mode using `$$$` and then enter the commands:

```
set uart baud 115200
save
reboot
```

The module will reboot into the new baud rate. Now to connect use the 115200 baud rate. So reconnect to the module using:

```sh
sudo minicom -D /dev/ttyAMA0 -b 115200
```

Now to set up ad hoc mode:

```
set ip dhcp 2
set wlan ssid <ssid>
set wlan chan <chan>
#set ip netmask 225.225.0.0
#set ip address 169.254.0.x # if needed
set wlan auth 0
set wlan join 4
```

Valid values for dhcp are:

- x=0: Static
- x=1: DHCP
- x=2: Auto (ad hoc)

Finally we set the required comm modes:

```
set ip proto 1
set ip host 255.255.255.255
set ip remote 8080
set ip local 8080
set comm match 4
set comm size 1524 #?
set comm time 0
save
reboot
```

## Protocol Buffers

Uses https://github.com/golang/protobuf for protocol buffer support. There is a Makefile in the packet.header package for building the protocol buffer specification found in that same package. To build this you will also need the protocol buffer package (Arch: `sudo pacman -S protobuf`).

Then we need the compiler plugin. From the README for golang/protobuf:

```sh
# Grab the code from the repository and install the proto package.
go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
```

The proto package will be pulled in as an SCmesh dependency.
