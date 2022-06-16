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
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/felixge/fgprof"

	"github.com/cloudwego/netpoll-benchmark/runner"
)

func Serve(newer runner.ServerNewer) {
	initFlags()
	// start pprof server
	go func() {
		http.DefaultServeMux.Handle("/debug/fgprof", fgprof.Handler())
		http.ListenAndServe(":18888", nil)
	}()

	svr := newer(runner.Mode(mode))
	err := svr.Run(network, address)
	if err != nil {
		log.Fatal(err)
	}
}

var (
	address string
	mode    int

	network string = "tcp"
)

func initFlags() {
	flag.StringVar(&address, "addr", ":9999", "client call address")
	flag.IntVar(&mode, "mode", 1, "1: echo, 2: idle, 3: mux")
	flag.Parse()
}
