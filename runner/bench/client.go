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

package bench

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/cloudwego/netpoll-benchmark/runner"
)

func Benching(newer runner.ClientNewer) {
	// start pprof server
	go func() {
		http.ListenAndServe(":19999", nil)
	}()

	initFlags()
	// prepare
	mod := runner.Mode(mode)
	client := newer(mod, network, address)
	var once runner.EchoOnce = client.Echo
	var conns = make([]runner.Conn, 4*concurrent)
	switch mod {
	case runner.Mode_Idle:
		// add idle conn 80%
		for i := range conns {
			conn, err := client.DialTimeout(network, address, dialTimeout)
			if err != nil {
				panic(fmt.Errorf("dial idle conn failed: %s", err.Error()))
			}
			conns[i] = conn
		}
	}
	// benching
	r := runner.NewRunner()
	r.Run(name, once, concurrent, total, echoSize)
}

var (
	name       string
	address    string
	echoSize   int
	total      int64
	concurrent int
	mode       int

	network     string        = "tcp"
	dialTimeout time.Duration = 50 * time.Millisecond
)

func initFlags() {
	flag.StringVar(&name, "name", "", "which repo to bench")
	flag.StringVar(&address, "addr", "127.0.0.1:9999", "client call address")
	flag.IntVar(&echoSize, "b", 1024, "echo size once")
	flag.IntVar(&concurrent, "c", 10, "call concurrent")
	flag.Int64Var(&total, "n", 100, "call total nums")
	flag.IntVar(&mode, "mode", 3, "1: echo, 2: idle, 3: mux")
	flag.Parse()
}
