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
	"bufio"
	"encoding/binary"
	"io"
	"sync"
	"unsafe"

	"github.com/cloudwego/netpoll-benchmark/runner"
)

const pagesize = 8192

var connerpool sync.Pool

type Conner struct {
	header []byte
	// conn   net.Conn
	rw *bufio.ReadWriter
}

func NewConner(conn io.ReadWriteCloser) *Conner {
	size := pagesize
	if v := connerpool.Get(); v != nil {
		conner := v.(*Conner)
		// conner.conn = conn
		conner.rw.Reader.Reset(conn)
		conner.rw.Writer.Reset(conn)
		return conner
	}
	conner := &Conner{}
	conner.header = make([]byte, 4)
	// conner.conn = conn
	conner.rw = bufio.NewReadWriter(bufio.NewReaderSize(conn, size), bufio.NewWriterSize(conn, size))
	return conner
}

func PutConner(conner *Conner) {
	conner.rw.Reader.Reset(nil)
	conner.rw.Writer.Reset(nil)
	connerpool.Put(conner)
}

// encode
func (c *Conner) Encode(msg *runner.Message) (err error) {
	binary.BigEndian.PutUint32(c.header, uint32(len(msg.Message)))
	_, err = c.rw.Write(c.header)
	if err == nil {
		_, err = c.rw.WriteString(msg.Message)
	}
	if err == nil {
		err = c.rw.Flush()
	}
	return err
}

// decode
func (c *Conner) Decode(msg *runner.Message) (err error) {
	// decode
	bLen, err := c.rw.Peek(4)
	if err == nil {
		_, err = c.rw.Discard(4)
	}
	if err != nil {
		return err
	}

	l := binary.BigEndian.Uint32(bLen)
	payload := make([]byte, l)
	_, err = io.ReadFull(c.rw, payload)
	if err != nil {
		return err
	}
	msg.Message = unsafeSliceToString(payload)
	return nil
}

// zero-copy slice convert to string
func unsafeSliceToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
