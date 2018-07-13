#!/bin/sh -e

name() {
	n="$(printf "%02d" "$1")"
	echo $n*
}

echo "Initializing xmc" >&2

set -x
total="$(ls ./scripts | wc -l)"
test ! -e /tasks && mkdir /tasks

last="$(ls /tasks | wc -l)"

if [ $last -eq $total ]; then
	echo "Already initialized. Exiting." >&2
	exit 0
fi

set +x

cd ./scripts
for i in $(seq $((last + 1)) $total); do
	echo "!!! Running $(echo $(name $i))" >&2
	source ./$(name $i) >&2
	:> /tasks/$i
done
