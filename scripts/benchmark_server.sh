#!/bin/bash

n=5000000
body=(1024) # 1KB
concurrent=(100 500 1000)

# no need to change
repo=("kcp")
ports=(7005)

. ./scripts/env.sh
. ./scripts/build.sh

# 1. echo
benchmark "server" $args
