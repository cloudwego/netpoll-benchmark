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
	"sync"

	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/pkg/pool/goroutine"

	"github.com/cloudwego/netpoll-benchmark/gnet/codec"
	"github.com/cloudwego/netpoll-benchmark/runner"
)

func NewMuxServer() runner.Server {
	return &muxServer{}
}

var _ runner.Server = &muxServer{}

type muxServer struct {
	*gnet.EventServer
	pool *goroutine.Pool
}

func (s *muxServer) Run(network, address string) error {
	p := goroutine.Default()
	defer p.Release()

	echo := &muxServer{pool: p}
	return gnet.Serve(echo, fmt.Sprintf("%s://%s", network, address), gnet.WithMulticore(true))
}

func (s *muxServer) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	mux := getOrNewMuxConn(c)
	for {
		req, err := mux.conner.Decode(frame)
		if err != nil {
			return nil, gnet.Close
		}
		if req == nil {
			return nil, gnet.None
		}
		frame = []byte{}

		// Use ants pool to unblock the event-loop.
		s.pool.Submit(func() {
			// handler
			resp := runner.ProcessRequest(reporter, req)
			// encode
			mux.wch <- resp
		})
	}
}

var muxconnmap sync.Map

func getOrNewMuxConn(conn gnet.Conn) *muxConn {
	c, loaded := muxconnmap.LoadOrStore(conn, &muxConn{})
	mux := c.(*muxConn)
	if loaded {
		return mux
	}
	// init
	mux.conn = conn
	mux.conner = codec.GetOrNewConner(conn)
	mux.wch = make(chan *runner.Message)
	go mux.loopWrite()
	return mux
}

type muxConn struct {
	conn   gnet.Conn
	conner *codec.Conner
	wch    chan *runner.Message
}

func (mux *muxConn) loopWrite() {
	for {
		msg := <-mux.wch
		// encode
		out := mux.conner.Encode(msg)
		err := mux.conn.AsyncWrite(out)
		if err != nil {
			panic(fmt.Errorf("mux encode failed: %s", err.Error()))
		}
	}
}
