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
	"net"
	"sync/atomic"
	"time"

	"github.com/xtaci/kcp-go"

	"github.com/cloudwego/netpoll-benchmark/net/codec"
	"github.com/cloudwego/netpoll-benchmark/runner"
)

func NewClientWithMux(network, address string, size int) runner.Client {
	cli := &muxclient{}
	cli.network = network
	cli.address = address
	cli.size = uint64(size)
	cli.conns = make([]*muxConn, size)
	for i := range cli.conns {
		cn, err := cli.DialTimeout(network, address, time.Second)
		if err != nil {
			panic(fmt.Errorf("mux dial conn failed: %s", err.Error()))
		}
		mc := newMuxConn(cn.(net.Conn))
		cli.conns[i] = mc
	}
	return cli
}

var _ runner.Client = &muxclient{}

type muxclient struct {
	network string
	address string
	conns   []*muxConn
	size    uint64
	cursor  uint64
}

func (cli *muxclient) DialTimeout(network, address string, timeout time.Duration) (runner.Conn, error) {
	return kcp.Dial(address)
}

func (cli *muxclient) Echo(req *runner.Message) (resp *runner.Message, err error) {
	// get conn & codec
	conn := cli.conns[atomic.AddUint64(&cli.cursor, 1)%cli.size]
	conn.wch <- req
	resp = <-conn.rch

	// reporter
	runner.ProcessResponse(resp)
	return resp, nil
}

func newMuxConn(conn net.Conn) *muxConn {
	mc := &muxConn{}
	mc.conn = conn
	mc.conner = codec.NewConner(conn)
	mc.rch = make(chan *runner.Message)
	mc.wch = make(chan *runner.Message)
	go mc.loopRead()
	go mc.loopWrite()
	return mc
}

type muxConn struct {
	conn     net.Conn
	conner   *codec.Conner
	rch, wch chan *runner.Message
}

func (mux *muxConn) loopRead() {
	for {
		msg := &runner.Message{}
		err := mux.conner.Decode(msg)
		if err != nil {
			panic(fmt.Errorf("mux decode failed: %s", err.Error()))
		}
		mux.rch <- msg
	}
}

func (mux *muxConn) loopWrite() {
	for {
		msg := <-mux.wch
		err := mux.conner.Encode(msg)
		if err != nil {
			panic(fmt.Errorf("mux encode failed: %s", err.Error()))
		}
	}
}
