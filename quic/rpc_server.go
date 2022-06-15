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
	"context"
	"strings"
	"time"

	"github.com/lucas-clemente/quic-go"

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
	listener, _ := quic.ListenAddr(address, generateTLSConfig(), nil)
	for {
		_conn, err := listener.Accept(context.Background())
		if err != nil {
			if strings.Contains(err.Error(), "closed") {
				return err
			}
			time.Sleep(10 * time.Millisecond) // too many open files ?
			continue
		}

		go func() {
			for {
				stream, err := _conn.AcceptStream(context.Background())
				if err != nil {
					panic(err)
				}
				go func() {
					conner := codec.NewConner(stream)
					defer codec.PutConner(conner)
					defer stream.Close()
					for err == nil {
						err = s.handler(conner)
					}
				}()
			}
		}()
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
	return conner.Encode(resp)
}
