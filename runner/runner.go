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
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/netpoll-benchmark/runner/perf"
)

type Runner struct {
	counter *Counter // 计数器
	timer   *Timer   // 计时器
}

func NewRunner() *Runner {
	r := &Runner{
		counter: NewCounter(),
		timer:   NewTimer(time.Microsecond),
	}
	return r
}

// 并发测试
func (r *Runner) Run(title string, once EchoOnce, concurrent int, total int64, echoSize int) {
	logInfo("%s start ..., concurrent: %d, total: %d", blueString("["+title+"]"), concurrent, total)

	// === warmup ===
	r.benching(once, concurrent, 10000, echoSize)

	// === starting ===
	if _, err := once(&Message{Message: ActionBegin}); err != nil {
		log.Fatalf("beginning server failed: %v", err)
	}
	recorder := perf.NewRecorder(fmt.Sprintf("%s@Client", strings.ToUpper(title)))
	recorder.Begin()

	// === benching ===
	start := r.timer.Now()
	r.benching(once, concurrent, total, echoSize)
	stop := r.timer.Now()
	r.counter.Report(title, stop-start, concurrent, total, echoSize)

	// == ending ===
	recorder.End()
	if _, err := once(&Message{Message: ActionEnd}); err != nil {
		log.Fatalf("ending server failed: %v", err)
	}

	// === reporting ===
	recorder.Report()
	fmt.Printf("\n\n")
}

func (r *Runner) benching(once EchoOnce, concurrent int, total int64, echoSize int) {
	var wg sync.WaitGroup
	wg.Add(concurrent)
	r.counter.Reset(total)

	// benching
	for i := 0; i < concurrent; i++ {
		go func() {
			defer wg.Done()

			ctx := context.Background()
			body := make([]byte, echoSize)
			req := &Message{Message: string(body)}
			for {
				idx := r.counter.Idx()
				if idx >= total {
					return
				}
				begin := r.timer.Now()
				err := r.echoWithTimeout(ctx, once, req)
				end := r.timer.Now()
				cost := end - begin
				r.counter.AddRecord(idx, err, cost)
			}
		}()
	}
	wg.Wait()
	r.counter.Total = total
}

// here is timeout wrapper which is necessary in rpc, timeout=1s.
func (r *Runner) echoWithTimeout(ctx context.Context, once EchoOnce, req *Message) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		_, err := once(req)
		done <- err
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return errors.New("echo timeout")
	}
}
