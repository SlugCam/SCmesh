# SCmesh

**Warning, this code is not suitable for anything at the moment. It is under heavy
development.**

[![Build Status](https://travis-ci.org/SlugCam/SCmesh.svg?branch=master)](https://travis-ci.org/SlugCam/SCmesh)
[![Coverage Status](https://coveralls.io/repos/SlugCam/SCmesh/badge.svg)](https://coveralls.io/r/SlugCam/SCmesh)

SCmesh is a daemon that implements mesh networking over a serial based wireless module (such as the WiFly). It is used to support an ad hoc networking mode in the SlugCam system.

## Protocol Buffers

Uses https://github.com/golang/protobuf for protocol buffer support. There is a Makefile in the packet.header package for building the protocol buffer specification found in that same package. To build this you will also need the protocol buffer package (Arch: `sudo pacman -S protobuf`).

Then we need the compiler plugin. From the README for golang/protobuf:

```sh
# Grab the code from the repository and install the proto package.
go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
```

The proto package will be pulled in as an SCmesh dependency.
