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
	"log"
	"net"
	"strings"
	"time"

	"github.com/xtaci/kcp-go"

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
	listener, err := kcp.Listen(address)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		log.Printf("connected: %v", conn.RemoteAddr())
		if err != nil {
			if strings.Contains(err.Error(), "closed") {
				return err
			}
			log.Println("accept failed: ", err)
			time.Sleep(10 * time.Millisecond) // too many open files ?
			continue
		}
		sess := conn.(*kcp.UDPSession)
		sess.SetMtu(65535)
		sess.SetNoDelay(0, 1, 0, 1)
		sess.SetWriteDelay(false)
		sess.SetWindowSize(1024, 1024)
		mc := newMuxConn(conn)
		go mc.loopRead()
		go mc.loopWrite()
	}
}

func newMuxConn(conn net.Conn) *muxConn {
	mc := &muxConn{}
	mc.conn = conn
	mc.conner = codec.NewConner(conn)
	mc.wch = make(chan *runner.Message)
	return mc
}

type muxConn struct {
	conn   net.Conn
	conner *codec.Conner
	wch    chan *runner.Message
}

func (mux *muxConn) loopRead() {
	for {
		msg := &runner.Message{}
		err := mux.conner.Decode(msg)
		if err != nil {
			panic(fmt.Errorf("mux decode failed: %s", err.Error()))
		}
		log.Printf("recv: %v", len(msg.Message))
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
	for msg := range mux.wch {
		log.Printf("send: %v", len(msg.Message))
		err := mux.conner.Encode(msg)
		if err != nil {
			panic(fmt.Errorf("mux encode failed: %s", err.Error()))
		}
	}
}
