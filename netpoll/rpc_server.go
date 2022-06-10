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

	"github.com/cloudwego/netpoll"

	"github.com/cloudwego/netpoll-benchmark/net/codec"
	"github.com/cloudwego/netpoll-benchmark/runner"
)

func NewRPCServer() runner.Server {
	return &rpcServer{}
}

var _ runner.Server = &rpcServer{}

type rpcServer struct{}

func (s *rpcServer) Run(network, address string) error {
	// new listener
	listener, err := netpoll.CreateListener(network, address)
	if err != nil {
		panic(err)
	}

	// new server
	//op := netpoll.WithOnPrepare(func(connection netpoll.Connection) context.Context {
	//	ctx := context.Background()
	//	ctx = context.WithValue(ctx, ctxconnerkey, &conner{
	//		conner: codec.NewConner(connection),
	//	})
	//	return ctx
	//})
	op := netpoll.WithNewer(func(ctx context.Context, c netpoll.Connection) netpoll.Handler {
		return &donner{
			conner: codec.NewConner2(c),
		}
	})
	eventLoop, err := netpoll.NewEventLoop(nil, op)
	if err != nil {
		panic(err)
	}
	// start listen loop ...
	return eventLoop.Serve(listener)
}

//
//type connerkey int
//
//const ctxconnerkey connerkey = 1

type donner struct {
	conner *codec.Conner
}

func (s *donner) OnRequest(ctx context.Context, conn netpoll.Connection) (err error) {
	//conner := ctx.Value(ctxconnerkey).(*conner).conner

	// decode
	req := &runner.Message{}
	err = s.conner.Decode(req)
	if err != nil {
		return err
	}

	// handler
	resp := runner.ProcessRequest(reporter, req)

	// encode
	err = s.conner.Encode(resp)
	if err != nil {
		return err
	}
	return nil
}
