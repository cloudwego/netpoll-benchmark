#!/bin/bash

nprocs=$(getconf _NPROCESSORS_ONLN)
if [ $nprocs -lt 4 ]; then
  echo "Your environment should have at least 4 processors"
  exit 1
elif [ $nprocs -gt 20 ]; then
  nprocs=20
fi

scpu=$((nprocs > 16 ? 3 : nprocs / 4 - 1)) # max is 3(4 cpus)
taskset_server="taskset -c 0-$scpu"
taskset_client="taskset -c $((scpu + 1))-$((nprocs - 1))"

# benchmark, args1=server|client
function benchmark() {
  . ./scripts/kill_servers.sh

  mode=${2-1}
  echo "benchmark mode=$mode"
  for b in ${body[@]}; do
    for c in ${concurrent[@]}; do
      for ((i = 0; i < ${#repo[@]}; i++)); do
        rp=${repo[i]}
        addr="127.0.0.1:${ports[i]}"

        # server start
        svr=${rp}_reciever
        if [[ "$1" == "client" ]]; then
          svr="net_reciever"
        fi
        nohup $taskset_server ./output/bin/${svr} -addr="$addr" -mode=$mode >>output/log/nohup.log 2>&1 &
        sleep 1
        echo "server $svr running with $taskset_server"

        # run client
        cli=${rp}_bencher
        if [[ "$1" == "server" ]]; then
          cli="net_bencher"
        fi
        echo "client $cli running with $taskset_client"
        $taskset_client ./output/bin/${cli} -name="$rp" -addr="$addr" -mode=$mode -b=$b -c=$c -n=$n

        # stop server
        pid=$(ps -ef | grep $svr | grep -v grep | awk '{print $2}')
        disown $pid
        kill -9 $pid
        sleep 1
      done
    done
  done
}
