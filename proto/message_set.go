// Go support for Protocol Buffers - Google's data interchange format
//
// Copyright 2010 Google Inc.  All rights reserved.
// http://code.google.com/p/goprotobuf/
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//     * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package proto

/*
 * Support for message sets.
 */

import (
	"bytes"
	"errors"
	"reflect"
)

// ErrNoMessageTypeId occurs when a protocol buffer does not have a message type ID.
// A message type ID is required for storing a protocol buffer in a message set.
var ErrNoMessageTypeId = errors.New("proto does not have a message type ID")

// The first two types (_MessageSet_Item and MessageSet)
// model what the protocol compiler produces for the following protocol message:
//   message MessageSet {
//     repeated group Item = 1 {
//       required int32 type_id = 2;
//       required string message = 3;
//     };
//   }
// That is the MessageSet wire format. We can't use a proto to generate these
// because that would introduce a circular dependency between it and this package.
//
// When a proto1 proto has a field that looks like:
//   optional message<MessageSet> info = 3;
// the protocol compiler produces a field in the generated struct that looks like:
//   Info *_proto_.MessageSet  `protobuf:"bytes,3,opt,name=info"`
// The package is automatically inserted so there is no need for that proto file to
// import this package.

type _MessageSet_Item struct {
	TypeId  *int32 `protobuf:"varint,2,req,name=type_id"`
	Message []byte `protobuf:"bytes,3,req,name=message"`
}

type MessageSet struct {
	Item             []*_MessageSet_Item `protobuf:"group,1,rep"`
	XXX_unrecognized *bytes.Buffer
	// TODO: caching?
}

// Make sure MessageSet is a Message.
var _ Message = (*MessageSet)(nil)

// messageTypeIder is an interface satisfied by a protocol buffer type
// that may be stored in a MessageSet.
type messageTypeIder interface {
	MessageTypeId() int32
}

func (ms *MessageSet) find(pb Message) *_MessageSet_Item {
	mti, ok := pb.(messageTypeIder)
	if !ok {
		return nil
	}
	id := mti.MessageTypeId()
	for _, item := range ms.Item {
		if *item.TypeId == id {
			return item
		}
	}
	return nil
}

func (ms *MessageSet) Has(pb Message) bool {
	if ms.find(pb) != nil {
		return true
	}
	return false
}

func (ms *MessageSet) Unmarshal(pb Message) error {
	if item := ms.find(pb); item != nil {
		return Unmarshal(item.Message, pb)
	}
	if _, ok := pb.(messageTypeIder); !ok {
		return ErrNoMessageTypeId
	}
	return nil // TODO: return error instead?
}

func (ms *MessageSet) Marshal(pb Message) error {
	msg, err := Marshal(pb)
	if err != nil {
		return err
	}
	if item := ms.find(pb); item != nil {
		// reuse existing item
		item.Message = msg
		return nil
	}

	mti, ok := pb.(messageTypeIder)
	if !ok {
		return ErrWrongType // TODO: custom error?
	}

	mtid := mti.MessageTypeId()
	ms.Item = append(ms.Item, &_MessageSet_Item{
		TypeId:  &mtid,
		Message: msg,
	})
	return nil
}

func (ms *MessageSet) Reset()         { *ms = MessageSet{} }
func (ms *MessageSet) String() string { return CompactTextString(ms) }
func (*MessageSet) ProtoMessage()     {}

// Support for the message_set_wire_format message option.

func skipVarint(buf []byte) []byte {
	i := 0
	for ; buf[i]&0x80 != 0; i++ {
	}
	return buf[i+1:]
}

// MarshalMessageSet encodes the extension map represented by m in the message set wire format.
// It is called by generated Marshal methods on protocol buffer messages with the message_set_wire_format option.
func MarshalMessageSet(m map[int32]Extension) ([]byte, error) {
	if err := encodeExtensionMap(m); err != nil {
		return nil, err
	}

	ms := &MessageSet{Item: make([]*_MessageSet_Item, len(m))}
	i := 0
	for k, e := range m {
		// Remove the wire type and field number varint, as well as the length varint.
		msg := skipVarint(skipVarint(e.enc))

		ms.Item[i] = &_MessageSet_Item{
			TypeId:  Int32(k),
			Message: msg,
		}
		i++
	}
	return Marshal(ms)
}

// UnmarshalMessageSet decodes the extension map encoded in buf in the message set wire format.
// It is called by generated Unmarshal methods on protocol buffer messages with the message_set_wire_format option.
func UnmarshalMessageSet(buf []byte, m map[int32]Extension) error {
	ms := new(MessageSet)
	if err := Unmarshal(buf, ms); err != nil {
		return err
	}
	for _, item := range ms.Item {
		// restore wire type and field number varint, plus length varint.
		b := EncodeVarint(uint64(*item.TypeId)<<3 | WireBytes)
		b = append(b, EncodeVarint(uint64(len(item.Message)))...)
		b = append(b, item.Message...)

		m[*item.TypeId] = Extension{enc: b}
	}
	return nil
}

// A global registry of types that can be used in a MessageSet.

var messageSetMap = make(map[int32]messageSetDesc)

type messageSetDesc struct {
	t    reflect.Type // pointer to struct
	name string
}

// RegisterMessageSetType is called from the generated code.
func RegisterMessageSetType(i messageTypeIder, name string) {
	messageSetMap[i.MessageTypeId()] = messageSetDesc{
		t:    reflect.TypeOf(i),
		name: name,
	}
}
