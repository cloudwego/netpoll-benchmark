#!/bin/bash

ps -ef | grep reciever | grep -v grep
pid=$(ps -ef | grep reciever | grep -v grep | awk '{print $2}')

# kill if server exist
if [[ -n "$pid" ]]; then
  kill -9 $pid
fi
