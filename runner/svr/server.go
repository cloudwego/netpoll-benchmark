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

package svr

import (
	"flag"
	"net/http"

	"github.com/cloudwego/netpoll-benchmark/runner"
)

func Serve(newer runner.ServerNewer) {
	initFlags()
	// start pprof server
	go func() {
		http.ListenAndServe("127.0.0.1:18888", nil)
	}()

	svr := newer(runner.Mode(mode))
	svr.Run(network, address)
}

var (
	address string
	mode    int

	network string = "tcp"
)

func initFlags() {
	flag.StringVar(&address, "addr", "127.0.0.1:9999", "client call address")
	flag.IntVar(&mode, "mode", 1, "1: echo, 2: idle, 3: mux")
	flag.Parse()
}
