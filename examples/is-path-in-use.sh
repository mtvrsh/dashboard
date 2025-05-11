#!/bin/sh
# $0 <PATH>... - exit code is 1 if any PATH is NOT in lsof output

if [ $# -lt 1 ]; then
	echo "error: provide at least 1 path"
	exit 2
fi

rg_cmd="rg -q"
for pattern in "$@"; do
	rg_cmd="${rg_cmd} -e ${pattern}"
done

lsof 2> /dev/null | $rg_cmd
