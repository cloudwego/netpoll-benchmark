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
	"fmt"
	"strings"

	"github.com/cloudwego/kitex-benchmark/perf"
)

const (
	MaxActionSize = 6
	ActionBegin   = "begin"
	ActionEnd     = "end"
	ActionReport  = "report"
)

func ProcessRequest(report *perf.Recorder, req *Message) (resp *Message) {
	if len(req.Message) < MaxActionSize {
		switch req.Message {
		case ActionBegin:
			report.Begin()
		case ActionEnd:
			report.End()
			return &Message{
				Message: ActionReport + report.ReportString(),
			}
		}
	}
	return &Message{
		Message: req.Message,
	}
}

func ProcessResponse(resp *Message) {
	if strings.HasPrefix(resp.Message, ActionReport) {
		fmt.Print(resp.Message[len(ActionReport):])
	}
}
