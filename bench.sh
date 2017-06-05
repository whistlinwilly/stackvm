#!/bin/bash
set -e

name=$1
if [ -z "$name" ]; then
    echo "missing name argument" >&2
    exit 1
fi

desc=$(git describe --always HEAD)

if ref="$(git symbolic-ref HEAD 2>/dev/null)"; then
    desc="$desc:${ref##*/}"
fi

out="prof.$name.$desc"
if [ -n "$2" ]; then
    out+=".$2"
fi

go test . -v -run "$name" -bench "$name" -benchmem -cpuprofile "$out.cpu" -memprofile "$out.mem" | tee "$out.bench"
