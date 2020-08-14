# Semgo

## Description

Because of the release cycle of machine image, in Semaphore, there is no guarantee that you are using the latest version of Go.

This can be a problem when Go has a CVE.

This tool is used to replace the version used by the command `sem-version`.

## Example

- Semaphore have only the `go1.14.6` and you want the latest `go1.14` (`go1.14.7`):

```console
$ sudo semgo go1.14
1.14.6 has been replaced by go1.14.7.

$ sem-version go 1.14

[18:29 14/08/2020]: Changing 'go' to version 1.14
Currently active Go version is:
go version go1.14.7 linux/amd64

[18:29 14/08/2020]: Switch successful.
```

- Semaphore have only the `go1.14.6` and you want the latest `go1.15` (`go1.15`):

```console
$ sudo semgo go1.15
1.14.6 has been replaced by go1.15.

$ sem-version go 1.14

[18:30 14/08/2020]: Changing 'go' to version 1.14
Currently active Go version is:
go version go1.15 linux/amd64

[18:30 14/08/2020]: Switch successful.
```
