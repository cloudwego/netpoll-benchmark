#!/bin/bash

# clean
rm -rf output/ && mkdir -p output/bin/ && mkdir -p output/log/

# build client
go build -v -o output/bin/net_bencher ./net/client
go build -v -o output/bin/netpoll_bencher ./netpoll/client
go build -v -o output/bin/gonet_benchmark ./gonet/client

# build server
go build -v -o output/bin/net_reciever ./net
go build -v -o output/bin/netpoll_reciever ./netpoll
go build -v -o output/bin/gnet_reciever ./gnet
go build -v -o output/bin/evio_reciever ./evio
go build -v -o output/bin/gonet_reciever ./gonet
