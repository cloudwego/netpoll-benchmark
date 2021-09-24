#!/bin/bash

. ./scripts/env.sh
. ./scripts/build.sh
. ./scripts/kill_servers.sh

core=0
for ((i = 0; i < ${#repo[@]}; i++)); do
  rp=${repo[i]}
  addr=":${ports[i]}"

  # server start
  nohup taskset -c $core-$(($core + 3)) ./output/bin/${rp}_reciever -addr="$addr" >>output/log/${rp}.log 2>&1 &
  echo "server $rp running at cpu $core-$(($core + 3))"
  core=$(($core + 4))
  sleep 1
done
