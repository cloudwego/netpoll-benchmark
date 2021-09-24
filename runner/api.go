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

package runner

import (
	"time"
)

/* Payload Format
 * Header: 4 bytes(indicate payload length)
 * Body:   length bytes
 */
type Message struct {
	Message string
}

type Client interface {
	DialTimeout(network, address string, timeout time.Duration) (conn Conn, err error)
	Echo(req *Message) (resp *Message, err error)
}

type ClientNewer func(mode Mode, network, address string) Client

type Server interface {
	Run(network, address string) error
}

type ServerNewer func(mode Mode) Server

// 屏蔽各 conn 定义的差异
type Conn interface {
	Close() error
}

type EchoOnce func(req *Message) (resp *Message, err error)

/*
 * Transport Mode:
 * 1. simple cho (connpool)
 * 2. long connpool + idle conns
 * 3. conn multiplexing
 */
type Mode int

const (
	Mode_Echo Mode = 1
	Mode_Idle Mode = 2 // 50% idle conn
	Mode_Mux  Mode = 3
)
