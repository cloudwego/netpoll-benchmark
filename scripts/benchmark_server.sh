#!/bin/bash

n=5000000
body=(1024) # 1KB
concurrent=(100 500 1000)

# no need to change
repo=("net" "netpoll" "gnet" "evio" "kcp")
ports=(7001 7002 7003 7004 7005)

. ./scripts/env.sh
. ./scripts/build.sh

# check args
args=${1-1}
if [[ "$args" == "3" ]]; then
  repo=("net" "netpoll" "gnet")
fi

# 1. echo
benchmark "server" $args
