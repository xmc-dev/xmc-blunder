#!/bin/sh

. "$(dirname "$0")/vars.sh"
cd "$XMC_DIR"

gen() {
	pushd proto

	for d in *; do
		pushd "$d"
		set -e
		protoc -I$GOPATH/src --go_out=plugins=micro:$GOPATH/src $PWD/*.proto
		set +e
		popd
	done
	popd
}

for s in $SERVICES; do
	pushd $s
	test -d proto && gen
	popd
done
