#!/bin/bash
set -e

name=$1
if [ -z "$name" ]; then
    echo "missing name argument" >&2
    exit 1
fi

if ref="$(git symbolic-ref HEAD 2>/dev/null)"; then
    ref=$(basename "$ref")
else
    ref=$(git describe --always HEAD)
fi

out="prof.$name.$ref"
if [ -n "$2" ]; then
    out+=".$2"
fi

go test . -v -run "$name" -bench "$name" -benchmem -cpuprofile "$out.cpu" -memprofile "$out.mem" | tee "$out.bench"
