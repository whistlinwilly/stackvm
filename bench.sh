#!/bin/bash
set -e

name=$1
branch=$(basename "$(git symbolic-ref HEAD)")
out="prof.$name.$branch"
if [ -n "$2" ]; then
    out+=".$2"
fi

go test . -v -run "$name" -bench "$name" -benchmem -cpuprofile "$out.cpu" -memprofile "$out.mem" | tee "$out.bench"
