#!/bin/bash

n=5000000
body=(1024) # 1KB
concurrent=(100 500 1000)

# no need to change
repo=("net" "netpoll")
ports=(7001 7002)

. ./scripts/env.sh
. ./scripts/build.sh

# echo
args=${1-1}
benchmark "client" $args
