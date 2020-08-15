# Semgo

[![GitHub release](https://img.shields.io/github/release/ldez/semgo.svg)](https://github.com/ldez/semgo/releases/latest)
[![Build Status](https://travis-ci.com/ldez/semgo.svg?branch=master)](https://travis-ci.com/ldez/semgo)

[![Sponsor me](https://img.shields.io/badge/Sponsor%20me-%E2%9D%A4%EF%B8%8F-pink.svg)](https://github.com/sponsors/ldez)

## Description

Because of the release cycle of machine image, in Semaphore, there is no guarantee that you are using the latest version of Go.

This can be a problem when Go has a CVE.

This tool is used to replace the version used by the command `sem-version`.

## Example

- Semaphore have only the `go1.14.6` and you want the latest `go1.14` (`go1.14.7`):

```console
$ sem-version go 1.14

[18:29 14/08/2020]: Changing 'go' to version 1.14
Currently active Go version is:
go version go1.14.6 linux/amd64

[18:29 14/08/2020]: Switch successful.

$ sudo semgo go1.14
[remote] go1.14.6 has been replaced by go1.14.7.
```

- Semaphore have only the `go1.14.6` and you want the latest `go1.15` (`go1.15`):

```console
$ sudo semgo go1.15
[remote] go1.10.8 has been replaced by go1.15.

```

## Installation

```bash
curl -sSfL https://raw.githubusercontent.com/ldez/semgo/master/godownloader.sh | sudo sh -s -- -b "/usr/local/bin"
```

```bash
curl -sSfL https://raw.githubusercontent.com/ldez/semgo/master/godownloader.sh | sudo sh -s -- -b "/usr/local/bin" v0.1.0
```

```bash
curl -sSfL https://raw.githubusercontent.com/ldez/semgo/master/godownloader.sh | sudo sh -s -- -b "/usr/local/bin" ${SEMGO_VERSION}
```
