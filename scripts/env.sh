#!/bin/bash

# cpu binding
nprocs=$(getconf _NPROCESSORS_ONLN)
if [ $nprocs -lt 4 ]; then
  echo "Your environment should have at least 4 processors"
  exit 1
elif [ $nprocs -gt 20 ]; then
  nprocs=20
fi
scpu=$((nprocs > 16 ? 4 : nprocs / 4)) # max is 4 cpus
ccpu=$((nprocs-scpu))
scpu_cmd="taskset -c 0-$((scpu-1))"
ccpu_cmd="taskset -c ${scpu}-$((ccpu-1))"
if [ -x "$(command -v numactl)" ]; then
  # use numa affinity
  node0=$(numactl -H | grep "node 0" | head -n 1 | cut -f "4-$((3+scpu))" -d ' ' --output-delimiter ',')
  node1=$(numactl -H | grep "node 1" | head -n 1 | cut -f "4-$((3+ccpu))" -d ' ' --output-delimiter ',')
  scpu_cmd="numactl -C ${node0} -m 0"
  ccpu_cmd="numactl -C ${node1} -m 1"
fi

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
        nohup $scpu_cmd ./output/bin/${svr} -addr="$addr" -mode=$mode >>output/log/nohup.log 2>&1 &
        sleep 1
        echo "server $svr running with $scpu_cmd"

        # run client
        cli=${rp}_bencher
        if [[ "$1" == "server" ]]; then
          cli="net_bencher"
        fi
        echo "client $cli running with $ccpu_cmd"
        $ccpu_cmd ./output/bin/${cli} -name="$rp" -addr="$addr" -mode=$mode -b=$b -c=$c -n=$n

        # stop server
        pid=$(ps -ef | grep $svr | grep -v grep | awk '{print $2}')
        if [[ -n "$pid" ]]; then
            disown $pid
            kill -9 $pid
        fi

        sleep 1
      done
    done
  done
}
