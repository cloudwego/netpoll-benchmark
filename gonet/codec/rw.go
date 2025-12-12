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
	"reflect"
	"unsafe"

	"github.com/bytedance/gopkg/lang/dirtmake"
	"github.com/cloudwego/gopkg/bufiox"

	"github.com/cloudwego/netpoll-benchmark/runner"
)

// encode
func Encode(writer bufiox.Writer, msg *runner.Message) (err error) {
	header, _ := writer.Malloc(4)
	binary.BigEndian.PutUint32(header, uint32(len(msg.Message)))

	_, _ = writer.WriteBinary(unsafeStringToSlice(msg.Message))
	err = writer.Flush()
	return err
}

// decode
func Decode(reader bufiox.Reader, msg *runner.Message) (err error) {
	bLen, err := reader.Next(4)
	if err != nil {
		return err
	}
	l := int(binary.BigEndian.Uint32(bLen))

	tmp := dirtmake.Bytes(l, l)

	_, err = reader.ReadBinary(tmp)
	if err != nil {
		return err
	}
	msg.Message = unsafeSliceToString(tmp)
	err = reader.Release(nil)
	return err
}

// zero-copy slice convert to string
func unsafeSliceToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// zero-copy slice convert to string
func unsafeStringToSlice(s string) (b []byte) {
	p := unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&s)).Data)
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	hdr.Data = uintptr(p)
	hdr.Cap = len(s)
	hdr.Len = len(s)
	return b
}
