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
	"log"
	"net"
	"strings"
	"time"

	"github.com/cloudwego/gopkg/bufiox"

	"github.com/cloudwego/netpoll-benchmark/gonet/codec"
	"github.com/cloudwego/netpoll-benchmark/runner"
)

func NewRPCServer() runner.Server {
	return &rpcServer{}
}

var _ runner.Server = &rpcServer{}

type rpcServer struct{}

func (s *rpcServer) Run(network, address string) error {
	// new listener
	listener, err := net.Listen(network, address)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "closed") {
				return err
			}
			time.Sleep(10 * time.Millisecond) // too many open files ?
			continue
		}

		go func() {
			defer conn.Close()
			reader, writer := bufiox.NewDefaultReader(conn), bufiox.NewDefaultWriter(conn)
			for {
				err := s.handler(context.Background(), reader, writer)
				if err != nil {
					log.Printf("rpcServer handler error: %v", err)
				}
			}
		}()
	}
}

func (s *rpcServer) handler(ctx context.Context, reader bufiox.Reader, writer bufiox.Writer) (err error) {
	// decode
	req := &runner.Message{}
	err = codec.Decode(reader, req)
	if err != nil {
		return err
	}

	// handler
	resp := runner.ProcessRequest(reporter, req)

	// encode
	err = codec.Encode(writer, resp)
	if err != nil {
		return err
	}
	return nil
}
