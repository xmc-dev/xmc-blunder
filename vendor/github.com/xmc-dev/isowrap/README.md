# Isowrap

[![Build Status](https://travis-ci.org/xmc-dev/isowrap.svg?branch=master)](https://travis-ci.org/xmc-dev/isowrap)
[![Coverage Status](https://coveralls.io/repos/github/xmc-dev/isowrap/badge.svg)](https://coveralls.io/github/xmc-dev/isowrap)
[![GoDoc](https://godoc.org/github.com/xmc-dev/isowrap?status.svg)](https://godoc.org/github.com/xmc-dev/isowrap)

Isowrap is a library used to execute programs isolated from the rest of the system.

It is a wrapper around Linux Containers (using [isolate](https://github.com/ioi/isolate)) and FreeBSD [jails](https://www.freebsd.org/doc/handbook/jails.html) (WIP).

This is probably alpha quality software.

## To do:

- [x] Linux isolate runner
  - [x] Full env
- [x] FreeBSD jail runner
  - [ ] DOES NOT COMPILE - breaking changes
  - [x] Implement "proper" wall time limit.
  - [x] Stack limit
  - [x] Maximum number of processes
  - [ ] Enable/Disable networking
  - [x] Environment variables

## Platform specific requirements

### Linux (`isolate`)

See the [INSTALLATION](https://github.com/ioi/isolate/blob/master/isolate.1.txt#L254-L280) part of the isolate manual. Control groups are required, make sure that they are enabled and `cgroupfs` is mounted.

### FreeBSD (`jail`)

Enable kernel `racct` support by adding the following line to `/etc/loader.conf`:

```
kern.racct.enable=1
```
