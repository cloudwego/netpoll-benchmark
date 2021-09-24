/*
 * Copyright 2021 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"fmt"
	"runtime"

	"github.com/tidwall/evio"

	"github.com/cloudwego/netpoll-benchmark/evio/codec"
	"github.com/cloudwego/netpoll-benchmark/runner"
)

func NewRPCServer() runner.Server {
	return &rpcServer{}
}

var _ runner.Server = &rpcServer{}

type rpcServer struct{}

func (s *rpcServer) Run(network, address string) error {
	events := &evio.Events{
		NumLoops:    runtime.GOMAXPROCS(0),
		LoadBalance: evio.RoundRobin,
		Data:        rpcHandler,
	}
	return evio.Serve(*events, fmt.Sprintf(`%s://%s`, network, address))
}

// evio 无法实现设定的测试场景
func rpcHandler(c evio.Conn, in []byte) (out []byte, action evio.Action) {
	conner := codec.GetOrNewConner(c)
	req, err := conner.Decode(in)
	if err != nil {
		return nil, evio.Close
	}
	if req == nil {
		return nil, evio.None
	}

	// 禁止串行处理
	var ch = make(chan int)
	go func() {
		// handler
		resp := runner.ProcessRequest(reporter, req)

		// encode
		out = conner.Encode(resp)
		action = evio.None
		ch <- 1
	}()
	<-ch
	return out, action
}
