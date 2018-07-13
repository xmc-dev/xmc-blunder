#!/bin/sh -e

echo "Installing micro cli" >&2
go get -v github.com/micro/micro

echo "Installing xmc-import" >&2
go get -v github.com/xmc-dev/xmc-core/cmd/xmc-import
