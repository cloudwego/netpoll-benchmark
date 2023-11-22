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
    "io"
	"net"
	"strings"
	"time"

	"github.com/cloudwego/netpoll-benchmark/net/codec"
	"github.com/cloudwego/netpoll-benchmark/runner"
)

func NewMuxServer() runner.Server {
	return &muxServer{}
}

var _ runner.Server = &muxServer{}

type muxServer struct{}

func (s *muxServer) Run(network, address string) error {
	// new listener
	listener, _ := net.Listen(network, address)
	var conns = make([]*muxConn, 0, 1024)
	for {
		_conn, err := listener.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "closed") {
				return err
			}
			time.Sleep(10 * time.Millisecond) // too many open files ?
			continue
		}

		mc := newMuxConn(_conn)
		conns = append(conns, mc)
	}
}

func newMuxConn(conn net.Conn) *muxConn {
	mc := &muxConn{}
	mc.conn = conn
	mc.conner = codec.NewConner(conn)
	mc.wch = make(chan *runner.Message)
    mc.ech = make(chan string)
	go mc.loopRead()
	go mc.loopWrite()
	return mc
}

type muxConn struct {
	conn   net.Conn
	conner *codec.Conner
	wch    chan *runner.Message
    ech    chan string
}

func (mux *muxConn) loopRead() {
	for {
		msg := &runner.Message{}
		err := mux.conner.Decode(msg)
		if err != nil {
            if err == io.EOF {
                mux.ech <- "EOF"
                // connection is closed.
                break;
            }
			panic(fmt.Errorf("mux decode failed: %s", err.Error()))
		}

		// handler must use another goroutine
		go func() {
			// handler
			resp := runner.ProcessRequest(reporter, msg)
			// encode
			mux.wch <- resp
		}()
	}
}

func (mux *muxConn) loopWrite() {
	for {
        select {
		    case msg := <-mux.wch:
		        err := mux.conner.Encode(msg)
		        if err != nil {
			        panic(fmt.Errorf("mux encode failed: %s", err.Error()))
                }
            case <-mux.ech:
                return
		}
	}
}

