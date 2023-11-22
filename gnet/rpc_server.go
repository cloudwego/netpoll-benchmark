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

	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/pkg/pool/goroutine"

	"github.com/cloudwego/netpoll-benchmark/gnet/codec"
	"github.com/cloudwego/netpoll-benchmark/runner"
)

func NewRPCServer() runner.Server {
	return &rpcServer{}
}

var _ runner.Server = &rpcServer{}

type rpcServer struct {
	*gnet.EventServer
	pool *goroutine.Pool
}

func (s *rpcServer) Run(network, address string) error {
	p := goroutine.Default()
	defer p.Release()

	echo := &rpcServer{pool: p}
	return gnet.Serve(echo, fmt.Sprintf("%s://%s", network, address), gnet.WithMulticore(true))
}

func (s *rpcServer) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	conner := codec.GetOrNewConner(c)
	req, err := conner.Decode(frame)
	if err != nil {
		return nil, gnet.Close
	}
	if req == nil {
		return nil, gnet.None
	}

	// Use ants pool to unblock the event-loop.
	s.pool.Submit(func() {
		// handler
		resp := runner.ProcessRequest(reporter, req)

		// encode
		out = conner.Encode(resp)
		err := c.AsyncWrite(out)
		if err != nil {
			panic(err)
		}
	})
	return
}
