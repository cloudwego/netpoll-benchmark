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

	"github.com/cloudwego/netpoll"

	"github.com/cloudwego/netpoll-benchmark/runner"
)

// encode
func Encode(writer netpoll.Writer, msg *runner.Message) (err error) {
	header, _ := writer.Malloc(4)
	binary.BigEndian.PutUint32(header, uint32(len(msg.Message)))

	writer.WriteString(msg.Message)
	err = writer.Flush()
	return err
}

// decode
func Decode(reader netpoll.Reader, msg *runner.Message) (err error) {
	bLen, err := reader.Next(4)
	if err != nil {
		return err
	}
	l := int(binary.BigEndian.Uint32(bLen))

	msg.Message, err = reader.ReadString(l)
	if err != nil {
		return err
	}
	err = reader.Release()
	return err
}
