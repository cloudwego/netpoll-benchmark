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

package connpool

import (
	"time"

	"github.com/cloudwego/netpoll-benchmark/runner"
)

func NewLongPool(maxIdle int) ConnPool {
	return &longpool{
		ring: NewRing(maxIdle),
	}
}

var _ ConnPool = &longpool{}

// Peer has one address, it manage all connections base on this address
type longpool struct {
	ring *Ring
}

// Get picks up connection from ring or dial a new one.
func (p *longpool) Get(network, address string, dialer Dialer, dialTimeout time.Duration) (runner.Conn, error) {
	for {
		conn, _ := p.ring.Pop().(runner.Conn)
		if conn == nil {
			break
		}
		return conn, nil
	}
	conn, err := dialer.DialTimeout(network, address, dialTimeout)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (p *longpool) Put(conn runner.Conn, err error) error {
	if err != nil {
		return conn.Close()
	}
	err = p.ring.Push(conn)
	if err != nil {
		return conn.Close()
	}
	return nil
}

// Close closes the peer and all the connections in the ring.
func (p *longpool) Close() error {
	for {
		conn, _ := p.ring.Pop().(runner.Conn)
		if conn == nil {
			break
		}
		conn.Close()
	}
	return nil
}
