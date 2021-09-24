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

package codec

import (
	"encoding/binary"
	"sync"

	"github.com/panjf2000/gnet"

	"github.com/cloudwego/netpoll-benchmark/runner"
)

var connermap sync.Map

type Conner struct {
	input []byte
	inlen int
}

func GetOrNewConner(conn gnet.Conn) *Conner {
	c, _ := connermap.LoadOrStore(conn, &Conner{})
	conner := c.(*Conner)
	return conner
}

// encode
func (c *Conner) Encode(msg *runner.Message) (out []byte) {
	out = make([]byte, 4+len(msg.Message))
	binary.BigEndian.PutUint32(out[:4], uint32(len(msg.Message)))
	copy(out[4:], msg.Message)
	return out
}

// decode
func (c *Conner) Decode(in []byte) (msg *runner.Message, err error) {
	c.input = append(c.input, in...)
	l := len(c.input)
	if c.inlen <= 0 {
		if l < 4 {
			return nil, nil
		}
		c.inlen = int(binary.BigEndian.Uint32(c.input[:4])) + 4
	}
	if l < c.inlen {
		return nil, nil
	}
	msg = &runner.Message{}
	msg.Message = string(c.input[4:c.inlen])
	// reset
	c.input = append(c.input[:0], c.input[c.inlen:]...)
	c.inlen = 0
	return msg, nil
}
