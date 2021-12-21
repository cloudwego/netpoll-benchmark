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
	"encoding/binary"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/cloudwego/netpoll"
	"github.com/cloudwego/netpoll/mux"

	"github.com/cloudwego/netpoll-benchmark/netpoll/codec"
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
		mc := newMuxConn(cn.(netpoll.Connection))
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
	return netpoll.DialConnection(network, address, timeout)
}

func (cli *muxclient) Echo(req *runner.Message) (resp *runner.Message, err error) {
	// get conn & codec
	mc := cli.conns[atomic.AddUint64(&cli.cursor, 1)%cli.size]

	// encode
	writer := netpoll.NewLinkBuffer()
	err = codec.Encode(writer, req)
	if err != nil {
		return nil, err
	}
	mc.Put(func() (buf netpoll.Writer, isNil bool) {
		return writer, false
	})

	// decode
	reader := <-mc.rch
	resp = &runner.Message{}
	err = codec.Decode(reader, resp)
	if err != nil {
		return nil, err
	}

	// reporter
	runner.ProcessResponse(resp)
	return resp, nil
}

func newMuxConn(conn netpoll.Connection) *muxConn {
	mc := &muxConn{}
	mc.conn = conn
	mc.rch = make(chan netpoll.Reader)
	// loop read
	conn.SetOnRequest(func(ctx context.Context, connection netpoll.Connection) error {
		reader := connection.Reader()
		// decode
		bLen, err := reader.Peek(4)
		if err != nil {
			return err
		}
		l := int(binary.BigEndian.Uint32(bLen))
		r, _ := reader.Slice(l + 4)
		mc.rch <- r
		return nil
	})
	// loop write
	mc.wqueue = mux.NewShardQueue(mux.ShardSize, conn)
	return mc
}

type muxConn struct {
	conn   netpoll.Connection
	rch    chan netpoll.Reader
	wqueue *mux.ShardQueue // use for write

}

// Put puts the buffer getter back to the queue.
func (c *muxConn) Put(gt mux.WriterGetter) {
	c.wqueue.Add(gt)
}
