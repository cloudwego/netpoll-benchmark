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
	"github.com/lucas-clemente/quic-go"

	"github.com/cloudwego/netpoll-benchmark/runner"
	"github.com/cloudwego/netpoll-benchmark/runner/bench"
)

// main is use for routing.
func main() {
	bench.Benching(NewClient)
}

func NewClient(mode runner.Mode, network, address string) runner.Client {
	switch mode {
	case runner.Mode_Echo, runner.Mode_Idle:
		return NewClientWithConnpool(network, address)
	case runner.Mode_Mux:
		return NewClientWithMux(network, address, 1)
	}
	return nil
}

type quicConn struct {
	quic.Connection
}

func (c *quicConn) Close() error {
	return c.CloseWithError(quic.ApplicationErrorCode(quic.NoError), "")
}
