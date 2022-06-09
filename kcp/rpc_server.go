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
	"net"

	"github.com/xtaci/kcp-go/v5"

	"github.com/cloudwego/netpoll-benchmark/net/codec"
	"github.com/cloudwego/netpoll-benchmark/runner"
)

func NewRPCServer() runner.Server {
	return &rpcServer{}
}

var _ runner.Server = &rpcServer{}

type rpcServer struct{}

func (s *rpcServer) Run(network, address string) error {
	// new listener
	listener, e := kcp.ListenWithOptions(address, nil, 10, 3)
	if e != nil {
		panic(e)
	}
	listener.SetReadBuffer(1024 * 1024)
	listener.SetWriteBuffer(1024 * 1024)
	listener.SetDSCP(46)

	for {
		_conn, err := listener.AcceptKCP()
		if err != nil {
			panic(err)
		}
		_conn.SetReadBuffer(1024 * 1024)
		_conn.SetWriteBuffer(1024 * 1024)

		go func(conn net.Conn) {
			conner := codec.NewConner(conn)
			defer codec.PutConner(conner)

			var err error
			for err == nil {
				err = s.handler(conner)
			}
			conn.Close()
		}(_conn)
	}
}

func (s *rpcServer) handler(conner *codec.Conner) (err error) {
	// decode
	req := &runner.Message{}
	err = conner.Decode(req)
	if err != nil {
		return err
	}

	// handler
	resp := runner.ProcessRequest(reporter, req)

	// encode
	err = conner.Encode(resp)
	if err != nil {
		return err
	}
	return nil
}
