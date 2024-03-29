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

package proto_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	. "./testdata"
	. "code.google.com/p/goprotobuf/proto"
)

var globalO *Buffer

func old() *Buffer {
	if globalO == nil {
		globalO = NewBuffer(nil)
	}
	globalO.Reset()
	return globalO
}

func equalbytes(b1, b2 []byte, t *testing.T) {
	if len(b1) != len(b2) {
		t.Errorf("wrong lengths: 2*%d != %d", len(b1), len(b2))
		return
	}
	for i := 0; i < len(b1); i++ {
		if b1[i] != b2[i] {
			t.Errorf("bad byte[%d]:%x %x: %s %s", i, b1[i], b2[i], b1, b2)
		}
	}
}

func initGoTestField() *GoTestField {
	f := new(GoTestField)
	f.Label = String("label")
	f.Type = String("type")
	return f
}

// These are all structurally equivalent but the tag numbers differ.
// (It's remarkable that required, optional, and repeated all have
// 8 letters.)
func initGoTest_RequiredGroup() *GoTest_RequiredGroup {
	return &GoTest_RequiredGroup{
		RequiredField: String("required"),
	}
}

func initGoTest_OptionalGroup() *GoTest_OptionalGroup {
	return &GoTest_OptionalGroup{
		RequiredField: String("optional"),
	}
}

func initGoTest_RepeatedGroup() *GoTest_RepeatedGroup {
	return &GoTest_RepeatedGroup{
		RequiredField: String("repeated"),
	}
}

func initGoTest(setdefaults bool) *GoTest {
	pb := new(GoTest)
	if setdefaults {
		pb.F_BoolDefaulted = Bool(Default_GoTest_F_BoolDefaulted)
		pb.F_Int32Defaulted = Int32(Default_GoTest_F_Int32Defaulted)
		pb.F_Int64Defaulted = Int64(Default_GoTest_F_Int64Defaulted)
		pb.F_Fixed32Defaulted = Uint32(Default_GoTest_F_Fixed32Defaulted)
		pb.F_Fixed64Defaulted = Uint64(Default_GoTest_F_Fixed64Defaulted)
		pb.F_Uint32Defaulted = Uint32(Default_GoTest_F_Uint32Defaulted)
		pb.F_Uint64Defaulted = Uint64(Default_GoTest_F_Uint64Defaulted)
		pb.F_FloatDefaulted = Float32(Default_GoTest_F_FloatDefaulted)
		pb.F_DoubleDefaulted = Float64(Default_GoTest_F_DoubleDefaulted)
		pb.F_StringDefaulted = String(Default_GoTest_F_StringDefaulted)
		pb.F_BytesDefaulted = Default_GoTest_F_BytesDefaulted
		pb.F_Sint32Defaulted = Int32(Default_GoTest_F_Sint32Defaulted)
		pb.F_Sint64Defaulted = Int64(Default_GoTest_F_Sint64Defaulted)
	}

	pb.Kind = GoTest_TIME.Enum()
	pb.RequiredField = initGoTestField()
	pb.F_BoolRequired = Bool(true)
	pb.F_Int32Required = Int32(3)
	pb.F_Int64Required = Int64(6)
	pb.F_Fixed32Required = Uint32(32)
	pb.F_Fixed64Required = Uint64(64)
	pb.F_Uint32Required = Uint32(3232)
	pb.F_Uint64Required = Uint64(6464)
	pb.F_FloatRequired = Float32(3232)
	pb.F_DoubleRequired = Float64(6464)
	pb.F_StringRequired = String("string")
	pb.F_BytesRequired = []byte("bytes")
	pb.F_Sint32Required = Int32(-32)
	pb.F_Sint64Required = Int64(-64)
	pb.Requiredgroup = initGoTest_RequiredGroup()

	return pb
}

func fail(msg string, b *bytes.Buffer, s string, t *testing.T) {
	data := b.Bytes()
	ld := len(data)
	ls := len(s) / 2

	fmt.Printf("fail %s ld=%d ls=%d\n", msg, ld, ls)

	// find the interesting spot - n
	n := ls
	if ld < ls {
		n = ld
	}
	j := 0
	for i := 0; i < n; i++ {
		bs := hex(s[j])*16 + hex(s[j+1])
		j += 2
		if data[i] == bs {
			continue
		}
		n = i
		break
	}
	l := n - 10
	if l < 0 {
		l = 0
	}
	h := n + 10

	// find the interesting spot - n
	fmt.Printf("is[%d]:", l)
	for i := l; i < h; i++ {
		if i >= ld {
			fmt.Printf(" --")
			continue
		}
		fmt.Printf(" %.2x", data[i])
	}
	fmt.Printf("\n")

	fmt.Printf("sb[%d]:", l)
	for i := l; i < h; i++ {
		if i >= ls {
			fmt.Printf(" --")
			continue
		}
		bs := hex(s[j])*16 + hex(s[j+1])
		j += 2
		fmt.Printf(" %.2x", bs)
	}
	fmt.Printf("\n")

	t.Fail()

	//	t.Errorf("%s: \ngood: %s\nbad: %x", msg, s, b.Bytes())
	// Print the output in a partially-decoded format; can
	// be helpful when updating the test.  It produces the output
	// that is pasted, with minor edits, into the argument to verify().
	//	data := b.Bytes()
	//	nesting := 0
	//	for b.Len() > 0 {
	//		start := len(data) - b.Len()
	//		var u uint64
	//		u, err := DecodeVarint(b)
	//		if err != nil {
	//			fmt.Printf("decode error on varint:", err)
	//			return
	//		}
	//		wire := u & 0x7
	//		tag := u >> 3
	//		switch wire {
	//		case WireVarint:
	//			v, err := DecodeVarint(b)
	//			if err != nil {
	//				fmt.Printf("decode error on varint:", err)
	//				return
	//			}
	//			fmt.Printf("\t\t\"%x\"  // field %d, encoding %d, value %d\n",
	//				data[start:len(data)-b.Len()], tag, wire, v)
	//		case WireFixed32:
	//			v, err := DecodeFixed32(b)
	//			if err != nil {
	//				fmt.Printf("decode error on fixed32:", err)
	//				return
	//			}
	//			fmt.Printf("\t\t\"%x\"  // field %d, encoding %d, value %d\n",
	//				data[start:len(data)-b.Len()], tag, wire, v)
	//		case WireFixed64:
	//			v, err := DecodeFixed64(b)
	//			if err != nil {
	//				fmt.Printf("decode error on fixed64:", err)
	//				return
	//			}
	//			fmt.Printf("\t\t\"%x\"  // field %d, encoding %d, value %d\n",
	//				data[start:len(data)-b.Len()], tag, wire, v)
	//		case WireBytes:
	//			nb, err := DecodeVarint(b)
	//			if err != nil {
	//				fmt.Printf("decode error on bytes:", err)
	//				return
	//			}
	//			after_tag := len(data) - b.Len()
	//			str := make([]byte, nb)
	//			_, err = b.Read(str)
	//			if err != nil {
	//				fmt.Printf("decode error on bytes:", err)
	//				return
	//			}
	//			fmt.Printf("\t\t\"%x\" \"%x\"  // field %d, encoding %d (FIELD)\n",
	//				data[start:after_tag], str, tag, wire)
	//		case WireStartGroup:
	//			nesting++
	//			fmt.Printf("\t\t\"%x\"\t\t// start group field %d level %d\n",
	//				data[start:len(data)-b.Len()], tag, nesting)
	//		case WireEndGroup:
	//			fmt.Printf("\t\t\"%x\"\t\t// end group field %d level %d\n",
	//				data[start:len(data)-b.Len()], tag, nesting)
	//			nesting--
	//		default:
	//			fmt.Printf("unrecognized wire type %d\n", wire)
	//			return
	//		}
	//	}
}

func hex(c uint8) uint8 {
	if '0' <= c && c <= '9' {
		return c - '0'
	}
	if 'a' <= c && c <= 'f' {
		return 10 + c - 'a'
	}
	if 'A' <= c && c <= 'F' {
		return 10 + c - 'A'
	}
	return 0
}

func equal(b []byte, s string, t *testing.T) bool {
	if 2*len(b) != len(s) {
		//		fail(fmt.Sprintf("wrong lengths: 2*%d != %d", len(b), len(s)), b, s, t)
		fmt.Printf("wrong lengths: 2*%d != %d\n", len(b), len(s))
		return false
	}
	for i, j := 0, 0; i < len(b); i, j = i+1, j+2 {
		x := hex(s[j])*16 + hex(s[j+1])
		if b[i] != x {
			//			fail(fmt.Sprintf("bad byte[%d]:%x %x", i, b[i], x), b, s, t)
			fmt.Printf("bad byte[%d]:%x %x", i, b[i], x)
			return false
		}
	}
	return true
}

func overify(t *testing.T, pb *GoTest, expected string) {
	o := old()
	err := o.Marshal(pb)
	if err != nil {
		fmt.Printf("overify marshal-1 err = %v", err)
		o.DebugPrint("", o.Bytes())
		t.Fatalf("expected = %s", expected)
	}
	if !equal(o.Bytes(), expected, t) {
		o.DebugPrint("overify neq 1", o.Bytes())
		t.Fatalf("expected = %s", expected)
	}

	// Now test Unmarshal by recreating the original buffer.
	pbd := new(GoTest)
	err = o.Unmarshal(pbd)
	if err != nil {
		t.Fatalf("overify unmarshal err = %v", err)
		o.DebugPrint("", o.Bytes())
		t.Fatalf("string = %s", expected)
	}
	o.Reset()
	err = o.Marshal(pbd)
	if err != nil {
		t.Errorf("overify marshal-2 err = %v", err)
		o.DebugPrint("", o.Bytes())
		t.Fatalf("string = %s", expected)
	}
	if !equal(o.Bytes(), expected, t) {
		o.DebugPrint("overify neq 2", o.Bytes())
		t.Fatalf("string = %s", expected)
	}
}

// Simple tests for numeric encode/decode primitives (varint, etc.)
func TestNumericPrimitives(t *testing.T) {
	for i := uint64(0); i < 1e6; i += 111 {
		o := old()
		if o.EncodeVarint(i) != nil {
			t.Error("EncodeVarint")
			break
		}
		x, e := o.DecodeVarint()
		if e != nil {
			t.Fatal("DecodeVarint")
		}
		if x != i {
			t.Fatal("varint decode fail:", i, x)
		}

		o = old()
		if o.EncodeFixed32(i) != nil {
			t.Fatal("encFixed32")
		}
		x, e = o.DecodeFixed32()
		if e != nil {
			t.Fatal("decFixed32")
		}
		if x != i {
			t.Fatal("fixed32 decode fail:", i, x)
		}

		o = old()
		if o.EncodeFixed64(i*1234567) != nil {
			t.Error("encFixed64")
			break
		}
		x, e = o.DecodeFixed64()
		if e != nil {
			t.Error("decFixed64")
			break
		}
		if x != i*1234567 {
			t.Error("fixed64 decode fail:", i*1234567, x)
			break
		}

		o = old()
		i32 := int32(i - 12345)
		if o.EncodeZigzag32(uint64(i32)) != nil {
			t.Fatal("EncodeZigzag32")
		}
		x, e = o.DecodeZigzag32()
		if e != nil {
			t.Fatal("DecodeZigzag32")
		}
		if x != uint64(uint32(i32)) {
			t.Fatal("zigzag32 decode fail:", i32, x)
		}

		o = old()
		i64 := int64(i - 12345)
		if o.EncodeZigzag64(uint64(i64)) != nil {
			t.Fatal("EncodeZigzag64")
		}
		x, e = o.DecodeZigzag64()
		if e != nil {
			t.Fatal("DecodeZigzag64")
		}
		if x != uint64(i64) {
			t.Fatal("zigzag64 decode fail:", i64, x)
		}
	}
}

// Simple tests for bytes
func TestBytesPrimitives(t *testing.T) {
	o := old()
	bytes := []byte{'n', 'o', 'w', ' ', 'i', 's', ' ', 't', 'h', 'e', ' ', 't', 'i', 'm', 'e'}
	if o.EncodeRawBytes(bytes) != nil {
		t.Error("EncodeRawBytes")
	}
	decb, e := o.DecodeRawBytes(false)
	if e != nil {
		t.Error("DecodeRawBytes")
	}
	equalbytes(bytes, decb, t)
}

// Simple tests for strings
func TestStringPrimitives(t *testing.T) {
	o := old()
	s := "now is the time"
	if o.EncodeStringBytes(s) != nil {
		t.Error("enc_string")
	}
	decs, e := o.DecodeStringBytes()
	if e != nil {
		t.Error("dec_string")
	}
	if s != decs {
		t.Error("string encode/decode fail:", s, decs)
	}
}

// Do we catch the "required bit not set" case?
func TestRequiredBit(t *testing.T) {
	o := old()
	pb := new(GoTest)
	err := o.Marshal(pb)
	if err == nil {
		t.Error("did not catch missing required fields")
	} else if strings.Index(err.Error(), "GoTest") < 0 {
		t.Error("wrong error type:", err)
	}
}

// Check that all fields are nil.
// Clearly silly, and a residue from a more interesting test with an earlier,
// different initialization property, but it once caught a compiler bug so
// it lives.
func checkInitialized(pb *GoTest, t *testing.T) {
	if pb.F_BoolDefaulted != nil {
		t.Error("New or Reset did not set boolean:", *pb.F_BoolDefaulted)
	}
	if pb.F_Int32Defaulted != nil {
		t.Error("New or Reset did not set int32:", *pb.F_Int32Defaulted)
	}
	if pb.F_Int64Defaulted != nil {
		t.Error("New or Reset did not set int64:", *pb.F_Int64Defaulted)
	}
	if pb.F_Fixed32Defaulted != nil {
		t.Error("New or Reset did not set fixed32:", *pb.F_Fixed32Defaulted)
	}
	if pb.F_Fixed64Defaulted != nil {
		t.Error("New or Reset did not set fixed64:", *pb.F_Fixed64Defaulted)
	}
	if pb.F_Uint32Defaulted != nil {
		t.Error("New or Reset did not set uint32:", *pb.F_Uint32Defaulted)
	}
	if pb.F_Uint64Defaulted != nil {
		t.Error("New or Reset did not set uint64:", *pb.F_Uint64Defaulted)
	}
	if pb.F_FloatDefaulted != nil {
		t.Error("New or Reset did not set float:", *pb.F_FloatDefaulted)
	}
	if pb.F_DoubleDefaulted != nil {
		t.Error("New or Reset did not set double:", *pb.F_DoubleDefaulted)
	}
	if pb.F_StringDefaulted != nil {
		t.Error("New or Reset did not set string:", *pb.F_StringDefaulted)
	}
	if pb.F_BytesDefaulted != nil {
		t.Error("New or Reset did not set bytes:", string(pb.F_BytesDefaulted))
	}
	if pb.F_Sint32Defaulted != nil {
		t.Error("New or Reset did not set int32:", *pb.F_Sint32Defaulted)
	}
	if pb.F_Sint64Defaulted != nil {
		t.Error("New or Reset did not set int64:", *pb.F_Sint64Defaulted)
	}
}

// Does Reset() reset?
func TestReset(t *testing.T) {
	pb := initGoTest(true)
	// muck with some values
	pb.F_BoolDefaulted = Bool(false)
	pb.F_Int32Defaulted = Int32(237)
	pb.F_Int64Defaulted = Int64(12346)
	pb.F_Fixed32Defaulted = Uint32(32000)
	pb.F_Fixed64Defaulted = Uint64(666)
	pb.F_Uint32Defaulted = Uint32(323232)
	pb.F_Uint64Defaulted = nil
	pb.F_FloatDefaulted = nil
	pb.F_DoubleDefaulted = Float64(0)
	pb.F_StringDefaulted = String("gotcha")
	pb.F_BytesDefaulted = []byte("asdfasdf")
	pb.F_Sint32Defaulted = Int32(123)
	pb.F_Sint64Defaulted = Int64(789)
	pb.Reset()
	checkInitialized(pb, t)
}

// All required fields set, no defaults provided.
func TestEncodeDecode1(t *testing.T) {
	pb := initGoTest(false)
	overify(t, pb,
		"0807"+ // field 1, encoding 0, value 7
			"220d"+"0a056c6162656c120474797065"+ // field 4, encoding 2 (GoTestField)
			"5001"+ // field 10, encoding 0, value 1
			"5803"+ // field 11, encoding 0, value 3
			"6006"+ // field 12, encoding 0, value 6
			"6d20000000"+ // field 13, encoding 5, value 0x20
			"714000000000000000"+ // field 14, encoding 1, value 0x40
			"78a019"+ // field 15, encoding 0, value 0xca0 = 3232
			"8001c032"+ // field 16, encoding 0, value 0x1940 = 6464
			"8d0100004a45"+ // field 17, encoding 5, value 3232.0
			"9101000000000040b940"+ // field 18, encoding 1, value 6464.0
			"9a0106"+"737472696e67"+ // field 19, encoding 2, string "string"
			"b304"+ // field 70, encoding 3, start group
			"ba0408"+"7265717569726564"+ // field 71, encoding 2, string "required"
			"b404"+ // field 70, encoding 4, end group
			"aa0605"+"6279746573"+ // field 101, encoding 2, string "bytes"
			"b0063f"+ // field 102, encoding 0, 0x3f zigzag32
			"b8067f") // field 103, encoding 0, 0x7f zigzag64
}

// All required fields set, defaults provided.
func TestEncodeDecode2(t *testing.T) {
	pb := initGoTest(true)
	overify(t, pb,
		"0807"+ // field 1, encoding 0, value 7
			"220d"+"0a056c6162656c120474797065"+ // field 4, encoding 2 (GoTestField)
			"5001"+ // field 10, encoding 0, value 1
			"5803"+ // field 11, encoding 0, value 3
			"6006"+ // field 12, encoding 0, value 6
			"6d20000000"+ // field 13, encoding 5, value 32
			"714000000000000000"+ // field 14, encoding 1, value 64
			"78a019"+ // field 15, encoding 0, value 3232
			"8001c032"+ // field 16, encoding 0, value 6464
			"8d0100004a45"+ // field 17, encoding 5, value 3232.0
			"9101000000000040b940"+ // field 18, encoding 1, value 6464.0
			"9a0106"+"737472696e67"+ // field 19, encoding 2 string "string"
			"c00201"+ // field 40, encoding 0, value 1
			"c80220"+ // field 41, encoding 0, value 32
			"d00240"+ // field 42, encoding 0, value 64
			"dd0240010000"+ // field 43, encoding 5, value 320
			"e1028002000000000000"+ // field 44, encoding 1, value 640
			"e8028019"+ // field 45, encoding 0, value 3200
			"f0028032"+ // field 46, encoding 0, value 6400
			"fd02e0659948"+ // field 47, encoding 5, value 314159.0
			"81030000000050971041"+ // field 48, encoding 1, value 271828.0
			"8a0310"+"68656c6c6f2c2022776f726c6421220a"+ // field 49, encoding 2 string "hello, \"world!\"\n"
			"b304"+ // start group field 70 level 1
			"ba0408"+"7265717569726564"+ // field 71, encoding 2, string "required"
			"b404"+ // end group field 70 level 1
			"aa0605"+"6279746573"+ // field 101, encoding 2 string "bytes"
			"b0063f"+ // field 102, encoding 0, 0x3f zigzag32
			"b8067f"+ // field 103, encoding 0, 0x7f zigzag64
			"8a1907"+"4269676e6f7365"+ // field 401, encoding 2, string "Bignose"
			"90193f"+ // field 402, encoding 0, value 63
			"98197f") // field 403, encoding 0, value 127

}

// All default fields set to their default value by hand
func TestEncodeDecode3(t *testing.T) {
	pb := initGoTest(false)
	pb.F_BoolDefaulted = Bool(true)
	pb.F_Int32Defaulted = Int32(32)
	pb.F_Int64Defaulted = Int64(64)
	pb.F_Fixed32Defaulted = Uint32(320)
	pb.F_Fixed64Defaulted = Uint64(640)
	pb.F_Uint32Defaulted = Uint32(3200)
	pb.F_Uint64Defaulted = Uint64(6400)
	pb.F_FloatDefaulted = Float32(314159)
	pb.F_DoubleDefaulted = Float64(271828)
	pb.F_StringDefaulted = String("hello, \"world!\"\n")
	pb.F_BytesDefaulted = []byte("Bignose")
	pb.F_Sint32Defaulted = Int32(-32)
	pb.F_Sint64Defaulted = Int64(-64)

	overify(t, pb,
		"0807"+ // field 1, encoding 0, value 7
			"220d"+"0a056c6162656c120474797065"+ // field 4, encoding 2 (GoTestField)
			"5001"+ // field 10, encoding 0, value 1
			"5803"+ // field 11, encoding 0, value 3
			"6006"+ // field 12, encoding 0, value 6
			"6d20000000"+ // field 13, encoding 5, value 32
			"714000000000000000"+ // field 14, encoding 1, value 64
			"78a019"+ // field 15, encoding 0, value 3232
			"8001c032"+ // field 16, encoding 0, value 6464
			"8d0100004a45"+ // field 17, encoding 5, value 3232.0
			"9101000000000040b940"+ // field 18, encoding 1, value 6464.0
			"9a0106"+"737472696e67"+ // field 19, encoding 2 string "string"
			"c00201"+ // field 40, encoding 0, value 1
			"c80220"+ // field 41, encoding 0, value 32
			"d00240"+ // field 42, encoding 0, value 64
			"dd0240010000"+ // field 43, encoding 5, value 320
			"e1028002000000000000"+ // field 44, encoding 1, value 640
			"e8028019"+ // field 45, encoding 0, value 3200
			"f0028032"+ // field 46, encoding 0, value 6400
			"fd02e0659948"+ // field 47, encoding 5, value 314159.0
			"81030000000050971041"+ // field 48, encoding 1, value 271828.0
			"8a0310"+"68656c6c6f2c2022776f726c6421220a"+ // field 49, encoding 2 string "hello, \"world!\"\n"
			"b304"+ // start group field 70 level 1
			"ba0408"+"7265717569726564"+ // field 71, encoding 2, string "required"
			"b404"+ // end group field 70 level 1
			"aa0605"+"6279746573"+ // field 101, encoding 2 string "bytes"
			"b0063f"+ // field 102, encoding 0, 0x3f zigzag32
			"b8067f"+ // field 103, encoding 0, 0x7f zigzag64
			"8a1907"+"4269676e6f7365"+ // field 401, encoding 2, string "Bignose"
			"90193f"+ // field 402, encoding 0, value 63
			"98197f") // field 403, encoding 0, value 127

}

// All required fields set, defaults provided, all non-defaulted optional fields have values.
func TestEncodeDecode4(t *testing.T) {
	pb := initGoTest(true)
	pb.Table = String("hello")
	pb.Param = Int32(7)
	pb.OptionalField = initGoTestField()
	pb.F_BoolOptional = Bool(true)
	pb.F_Int32Optional = Int32(32)
	pb.F_Int64Optional = Int64(64)
	pb.F_Fixed32Optional = Uint32(3232)
	pb.F_Fixed64Optional = Uint64(6464)
	pb.F_Uint32Optional = Uint32(323232)
	pb.F_Uint64Optional = Uint64(646464)
	pb.F_FloatOptional = Float32(32.)
	pb.F_DoubleOptional = Float64(64.)
	pb.F_StringOptional = String("hello")
	pb.F_BytesOptional = []byte("Bignose")
	pb.F_Sint32Optional = Int32(-32)
	pb.F_Sint64Optional = Int64(-64)
	pb.Optionalgroup = initGoTest_OptionalGroup()

	overify(t, pb,
		"0807"+ // field 1, encoding 0, value 7
			"1205"+"68656c6c6f"+ // field 2, encoding 2, string "hello"
			"1807"+ // field 3, encoding 0, value 7
			"220d"+"0a056c6162656c120474797065"+ // field 4, encoding 2 (GoTestField)
			"320d"+"0a056c6162656c120474797065"+ // field 6, encoding 2 (GoTestField)
			"5001"+ // field 10, encoding 0, value 1
			"5803"+ // field 11, encoding 0, value 3
			"6006"+ // field 12, encoding 0, value 6
			"6d20000000"+ // field 13, encoding 5, value 32
			"714000000000000000"+ // field 14, encoding 1, value 64
			"78a019"+ // field 15, encoding 0, value 3232
			"8001c032"+ // field 16, encoding 0, value 6464
			"8d0100004a45"+ // field 17, encoding 5, value 3232.0
			"9101000000000040b940"+ // field 18, encoding 1, value 6464.0
			"9a0106"+"737472696e67"+ // field 19, encoding 2 string "string"
			"f00101"+ // field 30, encoding 0, value 1
			"f80120"+ // field 31, encoding 0, value 32
			"800240"+ // field 32, encoding 0, value 64
			"8d02a00c0000"+ // field 33, encoding 5, value 3232
			"91024019000000000000"+ // field 34, encoding 1, value 6464
			"9802a0dd13"+ // field 35, encoding 0, value 323232
			"a002c0ba27"+ // field 36, encoding 0, value 646464
			"ad0200000042"+ // field 37, encoding 5, value 32.0
			"b1020000000000005040"+ // field 38, encoding 1, value 64.0
			"ba0205"+"68656c6c6f"+ // field 39, encoding 2, string "hello"
			"c00201"+ // field 40, encoding 0, value 1
			"c80220"+ // field 41, encoding 0, value 32
			"d00240"+ // field 42, encoding 0, value 64
			"dd0240010000"+ // field 43, encoding 5, value 320
			"e1028002000000000000"+ // field 44, encoding 1, value 640
			"e8028019"+ // field 45, encoding 0, value 3200
			"f0028032"+ // field 46, encoding 0, value 6400
			"fd02e0659948"+ // field 47, encoding 5, value 314159.0
			"81030000000050971041"+ // field 48, encoding 1, value 271828.0
			"8a0310"+"68656c6c6f2c2022776f726c6421220a"+ // field 49, encoding 2 string "hello, \"world!\"\n"
			"b304"+ // start group field 70 level 1
			"ba0408"+"7265717569726564"+ // field 71, encoding 2, string "required"
			"b404"+ // end group field 70 level 1
			"d305"+ // start group field 90 level 1
			"da0508"+"6f7074696f6e616c"+ // field 91, encoding 2, string "optional"
			"d405"+ // end group field 90 level 1
			"aa0605"+"6279746573"+ // field 101, encoding 2 string "bytes"
			"b0063f"+ // field 102, encoding 0, 0x3f zigzag32
			"b8067f"+ // field 103, encoding 0, 0x7f zigzag64
			"ea1207"+"4269676e6f7365"+ // field 301, encoding 2, string "Bignose"
			"f0123f"+ // field 302, encoding 0, value 63
			"f8127f"+ // field 303, encoding 0, value 127
			"8a1907"+"4269676e6f7365"+ // field 401, encoding 2, string "Bignose"
			"90193f"+ // field 402, encoding 0, value 63
			"98197f") // field 403, encoding 0, value 127

}

// All required fields set, defaults provided, all repeated fields given two values.
func TestEncodeDecode5(t *testing.T) {
	pb := initGoTest(true)
	pb.RepeatedField = []*GoTestField{initGoTestField(), initGoTestField()}
	pb.F_BoolRepeated = []bool{false, true}
	pb.F_Int32Repeated = []int32{32, 33}
	pb.F_Int64Repeated = []int64{64, 65}
	pb.F_Fixed32Repeated = []uint32{3232, 3333}
	pb.F_Fixed64Repeated = []uint64{6464, 6565}
	pb.F_Uint32Repeated = []uint32{323232, 333333}
	pb.F_Uint64Repeated = []uint64{646464, 656565}
	pb.F_FloatRepeated = []float32{32., 33.}
	pb.F_DoubleRepeated = []float64{64., 65.}
	pb.F_StringRepeated = []string{"hello", "sailor"}
	pb.F_BytesRepeated = [][]byte{[]byte("big"), []byte("nose")}
	pb.F_Sint32Repeated = []int32{32, -32}
	pb.F_Sint64Repeated = []int64{64, -64}
	pb.Repeatedgroup = []*GoTest_RepeatedGroup{initGoTest_RepeatedGroup(), initGoTest_RepeatedGroup()}

	overify(t, pb,
		"0807"+ // field 1, encoding 0, value 7
			"220d"+"0a056c6162656c120474797065"+ // field 4, encoding 2 (GoTestField)
			"2a0d"+"0a056c6162656c120474797065"+ // field 5, encoding 2 (GoTestField)
			"2a0d"+"0a056c6162656c120474797065"+ // field 5, encoding 2 (GoTestField)
			"5001"+ // field 10, encoding 0, value 1
			"5803"+ // field 11, encoding 0, value 3
			"6006"+ // field 12, encoding 0, value 6
			"6d20000000"+ // field 13, encoding 5, value 32
			"714000000000000000"+ // field 14, encoding 1, value 64
			"78a019"+ // field 15, encoding 0, value 3232
			"8001c032"+ // field 16, encoding 0, value 6464
			"8d0100004a45"+ // field 17, encoding 5, value 3232.0
			"9101000000000040b940"+ // field 18, encoding 1, value 6464.0
			"9a0106"+"737472696e67"+ // field 19, encoding 2 string "string"
			"a00100"+ // field 20, encoding 0, value 0
			"a00101"+ // field 20, encoding 0, value 1
			"a80120"+ // field 21, encoding 0, value 32
			"a80121"+ // field 21, encoding 0, value 33
			"b00140"+ // field 22, encoding 0, value 64
			"b00141"+ // field 22, encoding 0, value 65
			"bd01a00c0000"+ // field 23, encoding 5, value 3232
			"bd01050d0000"+ // field 23, encoding 5, value 3333
			"c1014019000000000000"+ // field 24, encoding 1, value 6464
			"c101a519000000000000"+ // field 24, encoding 1, value 6565
			"c801a0dd13"+ // field 25, encoding 0, value 323232
			"c80195ac14"+ // field 25, encoding 0, value 333333
			"d001c0ba27"+ // field 26, encoding 0, value 646464
			"d001b58928"+ // field 26, encoding 0, value 656565
			"dd0100000042"+ // field 27, encoding 5, value 32.0
			"dd0100000442"+ // field 27, encoding 5, value 33.0
			"e1010000000000005040"+ // field 28, encoding 1, value 64.0
			"e1010000000000405040"+ // field 28, encoding 1, value 65.0
			"ea0105"+"68656c6c6f"+ // field 29, encoding 2, string "hello"
			"ea0106"+"7361696c6f72"+ // field 29, encoding 2, string "sailor"
			"c00201"+ // field 40, encoding 0, value 1
			"c80220"+ // field 41, encoding 0, value 32
			"d00240"+ // field 42, encoding 0, value 64
			"dd0240010000"+ // field 43, encoding 5, value 320
			"e1028002000000000000"+ // field 44, encoding 1, value 640
			"e8028019"+ // field 45, encoding 0, value 3200
			"f0028032"+ // field 46, encoding 0, value 6400
			"fd02e0659948"+ // field 47, encoding 5, value 314159.0
			"81030000000050971041"+ // field 48, encoding 1, value 271828.0
			"8a0310"+"68656c6c6f2c2022776f726c6421220a"+ // field 49, encoding 2 string "hello, \"world!\"\n"
			"b304"+ // start group field 70 level 1
			"ba0408"+"7265717569726564"+ // field 71, encoding 2, string "required"
			"b404"+ // end group field 70 level 1
			"8305"+ // start group field 80 level 1
			"8a0508"+"7265706561746564"+ // field 81, encoding 2, string "repeated"
			"8405"+ // end group field 80 level 1
			"8305"+ // start group field 80 level 1
			"8a0508"+"7265706561746564"+ // field 81, encoding 2, string "repeated"
			"8405"+ // end group field 80 level 1
			"aa0605"+"6279746573"+ // field 101, encoding 2 string "bytes"
			"b0063f"+ // field 102, encoding 0, 0x3f zigzag32
			"b8067f"+ // field 103, encoding 0, 0x7f zigzag64
			"ca0c03"+"626967"+ // field 201, encoding 2, string "big"
			"ca0c04"+"6e6f7365"+ // field 201, encoding 2, string "nose"
			"d00c40"+ // field 202, encoding 0, value 32
			"d00c3f"+ // field 202, encoding 0, value -32
			"d80c8001"+ // field 203, encoding 0, value 64
			"d80c7f"+ // field 203, encoding 0, value -64
			"8a1907"+"4269676e6f7365"+ // field 401, encoding 2, string "Bignose"
			"90193f"+ // field 402, encoding 0, value 63
			"98197f") // field 403, encoding 0, value 127

}

// All required fields set, all packed repeated fields given two values.
func TestEncodeDecode6(t *testing.T) {
	pb := initGoTest(false)
	pb.F_BoolRepeatedPacked = []bool{false, true}
	pb.F_Int32RepeatedPacked = []int32{32, 33}
	pb.F_Int64RepeatedPacked = []int64{64, 65}
	pb.F_Fixed32RepeatedPacked = []uint32{3232, 3333}
	pb.F_Fixed64RepeatedPacked = []uint64{6464, 6565}
	pb.F_Uint32RepeatedPacked = []uint32{323232, 333333}
	pb.F_Uint64RepeatedPacked = []uint64{646464, 656565}
	pb.F_FloatRepeatedPacked = []float32{32., 33.}
	pb.F_DoubleRepeatedPacked = []float64{64., 65.}
	pb.F_Sint32RepeatedPacked = []int32{32, -32}
	pb.F_Sint64RepeatedPacked = []int64{64, -64}

	overify(t, pb,
		"0807"+ // field 1, encoding 0, value 7
			"220d"+"0a056c6162656c120474797065"+ // field 4, encoding 2 (GoTestField)
			"5001"+ // field 10, encoding 0, value 1
			"5803"+ // field 11, encoding 0, value 3
			"6006"+ // field 12, encoding 0, value 6
			"6d20000000"+ // field 13, encoding 5, value 32
			"714000000000000000"+ // field 14, encoding 1, value 64
			"78a019"+ // field 15, encoding 0, value 3232
			"8001c032"+ // field 16, encoding 0, value 6464
			"8d0100004a45"+ // field 17, encoding 5, value 3232.0
			"9101000000000040b940"+ // field 18, encoding 1, value 6464.0
			"9a0106"+"737472696e67"+ // field 19, encoding 2 string "string"
			"9203020001"+ // field 50, encoding 2, 2 bytes, value 0, value 1
			"9a03022021"+ // field 51, encoding 2, 2 bytes, value 32, value 33
			"a203024041"+ // field 52, encoding 2, 2 bytes, value 64, value 65
			"aa0308"+ // field 53, encoding 2, 8 bytes
			"a00c0000050d0000"+ // value 3232, value 3333
			"b20310"+ // field 54, encoding 2, 16 bytes
			"4019000000000000a519000000000000"+ // value 6464, value 6565
			"ba0306"+ // field 55, encoding 2, 6 bytes
			"a0dd1395ac14"+ // value 323232, value 333333
			"c20306"+ // field 56, encoding 2, 6 bytes
			"c0ba27b58928"+ // value 646464, value 656565
			"ca0308"+ // field 57, encoding 2, 8 bytes
			"0000004200000442"+ // value 32.0, value 33.0
			"d20310"+ // field 58, encoding 2, 16 bytes
			"00000000000050400000000000405040"+ // value 64.0, value 65.0
			"b304"+ // start group field 70 level 1
			"ba0408"+"7265717569726564"+ // field 71, encoding 2, string "required"
			"b404"+ // end group field 70 level 1
			"aa0605"+"6279746573"+ // field 101, encoding 2 string "bytes"
			"b0063f"+ // field 102, encoding 0, 0x3f zigzag32
			"b8067f"+ // field 103, encoding 0, 0x7f zigzag64
			"b21f02"+ // field 502, encoding 2, 2 bytes
			"403f"+ // value 32, value -32
			"ba1f03"+ // field 503, encoding 2, 3 bytes
			"80017f") // value 64, value -64
}

// Test that we can encode empty bytes fields.
func TestEncodeDecodeBytes1(t *testing.T) {
	pb := initGoTest(false)

	// Create our bytes
	pb.F_BytesRequired = []byte{}
	pb.F_BytesRepeated = [][]byte{{}}
	pb.F_BytesOptional = []byte{}

	d, err := Marshal(pb)
	if err != nil {
		t.Error(err)
	}

	pbd := new(GoTest)
	if err := Unmarshal(d, pbd); err != nil {
		t.Error(err)
	}

	if pbd.F_BytesRequired == nil || len(pbd.F_BytesRequired) != 0 {
		t.Error("required empty bytes field is incorrect")
	}
	if pbd.F_BytesRepeated == nil || len(pbd.F_BytesRepeated) == 1 && pbd.F_BytesRepeated[0] == nil {
		t.Error("repeated empty bytes field is incorrect")
	}
	if pbd.F_BytesOptional == nil || len(pbd.F_BytesOptional) != 0 {
		t.Error("optional empty bytes field is incorrect")
	}
}

// Test that we encode nil-valued fields of a repeated bytes field correctly.
// Since entries in a repeated field cannot be nil, nil must mean empty value.
func TestEncodeDecodeBytes2(t *testing.T) {
	pb := initGoTest(false)

	// Create our bytes
	pb.F_BytesRepeated = [][]byte{nil}

	d, err := Marshal(pb)
	if err != nil {
		t.Error(err)
	}

	pbd := new(GoTest)
	if err := Unmarshal(d, pbd); err != nil {
		t.Error(err)
	}

	if len(pbd.F_BytesRepeated) != 1 || pbd.F_BytesRepeated[0] == nil {
		t.Error("Unexpected value for repeated bytes field")
	}
}

// All required fields set, defaults provided, all repeated fields given two values.
func TestSkippingUnrecognizedFields(t *testing.T) {
	o := old()
	pb := initGoTestField()

	// Marshal it normally.
	o.Marshal(pb)

	// Now new a GoSkipTest record.
	skip := &GoSkipTest{
		SkipInt32:   Int32(32),
		SkipFixed32: Uint32(3232),
		SkipFixed64: Uint64(6464),
		SkipString:  String("skipper"),
		Skipgroup: &GoSkipTest_SkipGroup{
			GroupInt32:  Int32(75),
			GroupString: String("wxyz"),
		},
	}

	// Marshal it into same buffer.
	o.Marshal(skip)

	pbd := new(GoTestField)
	o.Unmarshal(pbd)

	// The __unrecognized field should be a marshaling of GoSkipTest
	skipd := new(GoSkipTest)

	o.SetBuf(pbd.XXX_unrecognized)
	o.Unmarshal(skipd)

	if *skipd.SkipInt32 != *skip.SkipInt32 {
		t.Error("skip int32", skipd.SkipInt32)
	}
	if *skipd.SkipFixed32 != *skip.SkipFixed32 {
		t.Error("skip fixed32", skipd.SkipFixed32)
	}
	if *skipd.SkipFixed64 != *skip.SkipFixed64 {
		t.Error("skip fixed64", skipd.SkipFixed64)
	}
	if *skipd.SkipString != *skip.SkipString {
		t.Error("skip string", *skipd.SkipString)
	}
	if *skipd.Skipgroup.GroupInt32 != *skip.Skipgroup.GroupInt32 {
		t.Error("skip group int32", skipd.Skipgroup.GroupInt32)
	}
	if *skipd.Skipgroup.GroupString != *skip.Skipgroup.GroupString {
		t.Error("skip group string", *skipd.Skipgroup.GroupString)
	}
}

// Check that we can grow an array (repeated field) to have many elements.
// This test doesn't depend only on our encoding; for variety, it makes sure
// we create, encode, and decode the correct contents explicitly.  It's therefore
// a bit messier.
// This test also uses (and hence tests) the Marshal/Unmarshal functions
// instead of the methods.
func TestBigRepeated(t *testing.T) {
	pb := initGoTest(true)

	// Create the arrays
	const N = 50 // Internally the library starts much smaller.
	pb.Repeatedgroup = make([]*GoTest_RepeatedGroup, N)
	pb.F_Sint64Repeated = make([]int64, N)
	pb.F_Sint32Repeated = make([]int32, N)
	pb.F_BytesRepeated = make([][]byte, N)
	pb.F_StringRepeated = make([]string, N)
	pb.F_DoubleRepeated = make([]float64, N)
	pb.F_FloatRepeated = make([]float32, N)
	pb.F_Uint64Repeated = make([]uint64, N)
	pb.F_Uint32Repeated = make([]uint32, N)
	pb.F_Fixed64Repeated = make([]uint64, N)
	pb.F_Fixed32Repeated = make([]uint32, N)
	pb.F_Int64Repeated = make([]int64, N)
	pb.F_Int32Repeated = make([]int32, N)
	pb.F_BoolRepeated = make([]bool, N)
	pb.RepeatedField = make([]*GoTestField, N)

	// Fill in the arrays with checkable values.
	igtf := initGoTestField()
	igtrg := initGoTest_RepeatedGroup()
	for i := 0; i < N; i++ {
		pb.Repeatedgroup[i] = igtrg
		pb.F_Sint64Repeated[i] = int64(i)
		pb.F_Sint32Repeated[i] = int32(i)
		s := fmt.Sprint(i)
		pb.F_BytesRepeated[i] = []byte(s)
		pb.F_StringRepeated[i] = s
		pb.F_DoubleRepeated[i] = float64(i)
		pb.F_FloatRepeated[i] = float32(i)
		pb.F_Uint64Repeated[i] = uint64(i)
		pb.F_Uint32Repeated[i] = uint32(i)
		pb.F_Fixed64Repeated[i] = uint64(i)
		pb.F_Fixed32Repeated[i] = uint32(i)
		pb.F_Int64Repeated[i] = int64(i)
		pb.F_Int32Repeated[i] = int32(i)
		pb.F_BoolRepeated[i] = i%2 == 0
		pb.RepeatedField[i] = igtf
	}

	// Marshal.
	buf, _ := Marshal(pb)

	// Now test Unmarshal by recreating the original buffer.
	pbd := new(GoTest)
	Unmarshal(buf, pbd)

	// Check the checkable values
	for i := uint64(0); i < N; i++ {
		if pbd.Repeatedgroup[i] == nil { // TODO: more checking?
			t.Error("pbd.Repeatedgroup bad")
		}
		var x uint64
		x = uint64(pbd.F_Sint64Repeated[i])
		if x != i {
			t.Error("pbd.F_Sint64Repeated bad", x, i)
		}
		x = uint64(pbd.F_Sint32Repeated[i])
		if x != i {
			t.Error("pbd.F_Sint32Repeated bad", x, i)
		}
		s := fmt.Sprint(i)
		equalbytes(pbd.F_BytesRepeated[i], []byte(s), t)
		if pbd.F_StringRepeated[i] != s {
			t.Error("pbd.F_Sint32Repeated bad", pbd.F_StringRepeated[i], i)
		}
		x = uint64(pbd.F_DoubleRepeated[i])
		if x != i {
			t.Error("pbd.F_DoubleRepeated bad", x, i)
		}
		x = uint64(pbd.F_FloatRepeated[i])
		if x != i {
			t.Error("pbd.F_FloatRepeated bad", x, i)
		}
		x = pbd.F_Uint64Repeated[i]
		if x != i {
			t.Error("pbd.F_Uint64Repeated bad", x, i)
		}
		x = uint64(pbd.F_Uint32Repeated[i])
		if x != i {
			t.Error("pbd.F_Uint32Repeated bad", x, i)
		}
		x = pbd.F_Fixed64Repeated[i]
		if x != i {
			t.Error("pbd.F_Fixed64Repeated bad", x, i)
		}
		x = uint64(pbd.F_Fixed32Repeated[i])
		if x != i {
			t.Error("pbd.F_Fixed32Repeated bad", x, i)
		}
		x = uint64(pbd.F_Int64Repeated[i])
		if x != i {
			t.Error("pbd.F_Int64Repeated bad", x, i)
		}
		x = uint64(pbd.F_Int32Repeated[i])
		if x != i {
			t.Error("pbd.F_Int32Repeated bad", x, i)
		}
		if pbd.F_BoolRepeated[i] != (i%2 == 0) {
			t.Error("pbd.F_BoolRepeated bad", x, i)
		}
		if pbd.RepeatedField[i] == nil { // TODO: more checking?
			t.Error("pbd.RepeatedField bad")
		}
	}
}

// Verify we give a useful message when decoding to the wrong structure type.
func TestTypeMismatch(t *testing.T) {
	pb1 := initGoTest(true)

	// Marshal
	o := old()
	o.Marshal(pb1)

	// Now Unmarshal it to the wrong type.
	pb2 := initGoTestField()
	err := o.Unmarshal(pb2)
	switch err {
	case ErrWrongType:
		// fine
	case nil:
		t.Error("expected wrong type error, got no error")
	default:
		t.Error("expected wrong type error, got", err)
	}
}

func encodeDecode(t *testing.T, in, out Message, msg string) {
	buf, err := Marshal(in)
	if err != nil {
		t.Fatalf("failed marshaling %v: %v", msg, err)
	}
	if err := Unmarshal(buf, out); err != nil {
		t.Fatalf("failed unmarshaling %v: %v", msg, err)
	}
}

func TestPackedNonPackedDecoderSwitching(t *testing.T) {
	np, p := new(NonPackedTest), new(PackedTest)

	// non-packed -> packed
	np.A = []int32{0, 1, 1, 2, 3, 5}
	encodeDecode(t, np, p, "non-packed -> packed")
	if !reflect.DeepEqual(np.A, p.B) {
		t.Errorf("failed non-packed -> packed; np.A=%+v, p.B=%+v", np.A, p.B)
	}

	// packed -> non-packed
	np.Reset()
	p.B = []int32{3, 1, 4, 1, 5, 9}
	encodeDecode(t, p, np, "packed -> non-packed")
	if !reflect.DeepEqual(p.B, np.A) {
		t.Errorf("failed packed -> non-packed; p.B=%+v, np.A=%+v", p.B, np.A)
	}
}

func TestProto1RepeatedGroup(t *testing.T) {
	pb := &MessageList{
		Message: []*MessageList_Message{
			&MessageList_Message{
				Name:  String("blah"),
				Count: Int32(7),
			},
			// NOTE: pb.Message[1] is a nil
			nil,
		},
	}

	o := old()
	if err := o.Marshal(pb); err != ErrRepeatedHasNil {
		t.Fatalf("unexpected or no error when marshaling: %v", err)
	}
}

// Test that enums work.  Checks for a bug introduced by making enums
// named types instead of int32: newInt32FromUint64 would crash with
// a type mismatch in reflect.PointTo.
func TestEnum(t *testing.T) {
	pb := new(GoEnum)
	pb.Foo = FOO_FOO1.Enum()
	o := old()
	if err := o.Marshal(pb); err != nil {
		t.Fatal("error encoding enum:", err)
	}
	pb1 := new(GoEnum)
	if err := o.Unmarshal(pb1); err != nil {
		t.Fatal("error decoding enum:", err)
	}
	if *pb1.Foo != FOO_FOO1 {
		t.Error("expected 7 but got ", *pb1.Foo)
	}
}

// Enum types have String methods. Check that enum fields can be printed.
// We don't care what the value actually is, just as long as it doesn't crash.
func TestPrintingNilEnumFields(t *testing.T) {
	pb := new(GoEnum)
	fmt.Sprintf("%+v", pb)
}

// Verify that absent required fields cause Marshal/Unmarshal to return errors.
func TestRequiredFieldEnforcement(t *testing.T) {
	pb := new(GoTestField)
	_, err := Marshal(pb)
	if err == nil {
		t.Error("marshal: expected error, got nil")
	} else if strings.Index(err.Error(), "GoTestField") < 0 {
		t.Errorf("marshal: bad error type: %v", err)
	}

	// A slightly sneaky, yet valid, proto. It encodes the same required field twice,
	// so simply counting the required fields is insufficient.
	// field 1, encoding 2, value "hi"
	buf := []byte("\x0A\x02hi\x0A\x02hi")
	err = Unmarshal(buf, pb)
	if err == nil {
		t.Error("unmarshal: expected error, got nil")
	} else if strings.Index(err.Error(), "GoTestField") < 0 {
		t.Errorf("unmarshal: bad error type: %v", err)
	}
}

// A type that implements the Marshaler interface, but is not nillable.
type nonNillableInt uint64

func (nni nonNillableInt) Marshal() ([]byte, error) {
	return EncodeVarint(uint64(nni)), nil
}

type NNIMessage struct {
	nni nonNillableInt
}

func (*NNIMessage) Reset()         {}
func (*NNIMessage) String() string { return "" }
func (*NNIMessage) ProtoMessage()  {}

// A type that implements the Marshaler interface and is nillable.
type nillableMessage struct {
	x uint64
}

func (nm *nillableMessage) Marshal() ([]byte, error) {
	return EncodeVarint(nm.x), nil
}

type NMMessage struct {
	nm *nillableMessage
}

func (*NMMessage) Reset()         {}
func (*NMMessage) String() string { return "" }
func (*NMMessage) ProtoMessage()  {}

// Verify a type that uses the Marshaler interface, but has a nil pointer.
func TestNilMarshaler(t *testing.T) {
	// Try a struct with a Marshaler field that is nil.
	// It should be directly marshable.
	nmm := new(NMMessage)
	if _, err := Marshal(nmm); err != nil {
		t.Error("unexpected error marshaling nmm: ", err)
	}

	// Try a struct with a Marshaler field that is not nillable.
	nnim := new(NNIMessage)
	nnim.nni = 7
	var _ Marshaler = nnim.nni // verify it is truly a Marshaler
	if _, err := Marshal(nnim); err != nil {
		t.Error("unexpected error marshaling nnim: ", err)
	}
}

func TestAllSetDefaults(t *testing.T) {
	// Exercise SetDefaults with all scalar field types.
	m := &Defaults{
		// NaN != NaN, so override that here.
		F_Nan: Float32(1.7),
	}
	expected := &Defaults{
		F_Bool:    Bool(true),
		F_Int32:   Int32(32),
		F_Int64:   Int64(64),
		F_Fixed32: Uint32(320),
		F_Fixed64: Uint64(640),
		F_Uint32:  Uint32(3200),
		F_Uint64:  Uint64(6400),
		F_Float:   Float32(314159),
		F_Double:  Float64(271828),
		F_String:  String(`hello, "world!"` + "\n"),
		F_Bytes:   []byte("Bignose"),
		F_Sint32:  Int32(-32),
		F_Sint64:  Int64(-64),
		F_Enum:    Defaults_GREEN.Enum(),
		F_Pinf:    Float32(float32(math.Inf(1))),
		F_Ninf:    Float32(float32(math.Inf(-1))),
		F_Nan:     Float32(1.7),
	}
	SetDefaults(m)
	if !Equal(m, expected) {
		t.Errorf(" got %v\nwant %v", m, expected)
	}
}

func TestSetDefaultsWithSetField(t *testing.T) {
	// Check that a set value is not overridden.
	m := &Defaults{
		F_Int32: Int32(12),
	}
	SetDefaults(m)
	if v := m.GetF_Int32(); v != 12 {
		t.Errorf("m.FInt32 = %v, want 12", v)
	}
}

func TestSetDefaultsWithSubMessage(t *testing.T) {
	m := &OtherMessage{
		Key: Int64(123),
		Inner: &InnerMessage{
			Host: String("gopher"),
		},
	}
	expected := &OtherMessage{
		Key: Int64(123),
		Inner: &InnerMessage{
			Host: String("gopher"),
			Port: Int32(4000),
		},
	}
	SetDefaults(m)
	if !Equal(m, expected) {
		t.Errorf("\n got %v\nwant %v", m, expected)
	}
}

func TestMaximumTagNumber(t *testing.T) {
	m := &MaxTag{
		LastField: String("natural goat essence"),
	}
	buf, err := Marshal(m)
	if err != nil {
		t.Fatalf("proto.Marshal failed: %v", err)
	}
	m2 := new(MaxTag)
	if err := Unmarshal(buf, m2); err != nil {
		t.Fatalf("proto.Unmarshal failed: %v", err)
	}
	if got, want := m2.GetLastField(), *m.LastField; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestJSON(t *testing.T) {
	m := &MyMessage{
		Count: Int32(4),
		Pet:   []string{"bunny", "kitty"},
		Inner: &InnerMessage{
			Host: String("cauchy"),
		},
	}
	const expected = `{"count":4,"pet":["bunny","kitty"],"inner":{"host":"cauchy"}}`

	b, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	s := string(b)
	if s != expected {
		t.Errorf("got  %s\nwant %s", s, expected)
	}
}

func TestBadWireType(t *testing.T) {
	b := []byte{7<<3 | 6} // field 7, wire type 6
	pb := new(OtherMessage)
	if err := Unmarshal(b, pb); err == nil {
		t.Errorf("Unmarshal did not fail")
	} else if !strings.Contains(err.Error(), "unknown wire type") {
		t.Errorf("wrong error: %v", err)
	}
}

func TestBytesWithInvalidLength(t *testing.T) {
	// If a byte sequence has an invalid (negative) length, Unmarshal should not panic.
	b := []byte{2<<3 | WireBytes, 0xff, 0xff, 0xff, 0xff, 0xff, 0}
	Unmarshal(b, new(MyMessage))
}

func TestUnmarshalFuzz(t *testing.T) {
	const N = 1000
	seed := time.Now().UnixNano()
	t.Logf("RNG seed is %d", seed)
	rng := rand.New(rand.NewSource(seed))
	buf := make([]byte, 20)
	for i := 0; i < N; i++ {
		for j := range buf {
			buf[j] = byte(rng.Intn(256))
		}
		fuzzUnmarshal(t, buf)
	}
}

func TestAppend(t *testing.T) {
	pb := &MessageList{Message: []*MessageList_Message{{Name: String("x"), Count: Int32(1)}}}
	data, err := Marshal(pb)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	pb1 := new(MessageList)
	if err := Unmarshal(data, pb1); err != nil {
		t.Fatalf("first Unmarshal: %v", err)
	}
	if err := Unmarshal(data, pb1); err != nil {
		t.Fatalf("second Unmarshal: %v", err)
	}
	if len(pb1.Message) != 1 {
		t.Errorf("two Unmarshals produced %d Messages, want 1", len(pb1.Message))
	}

	pb2 := new(MessageList)
	if err := UnmarshalAppend(data, pb2); err != nil {
		t.Fatalf("first UnmarshalAppend: %v", err)
	}
	if err := UnmarshalAppend(data, pb2); err != nil {
		t.Fatalf("second UnmarshalAppend: %v", err)
	}
	if len(pb2.Message) != 2 {
		t.Errorf("two UnmarshalAppends produced %d Messages, want 2", len(pb2.Message))
	}
}

func fuzzUnmarshal(t *testing.T, data []byte) {
	defer func() {
		if e := recover(); e != nil {
			t.Errorf("These bytes caused a panic: %+v", data)
			t.Logf("Stack:\n%s", debug.Stack())
			t.FailNow()
		}
	}()

	pb := new(MyMessage)
	Unmarshal(data, pb)
}

func BenchmarkMarshal(b *testing.B) {
	b.StopTimer()

	pb := initGoTest(true)

	// Create an array
	const N = 1000 // Internally the library starts much smaller.
	pb.F_Int32Repeated = make([]int32, N)
	pb.F_DoubleRepeated = make([]float64, N)

	// Fill in the array with some values.
	for i := 0; i < N; i++ {
		pb.F_Int32Repeated[i] = int32(i)
		pb.F_DoubleRepeated[i] = float64(i)
	}

	p := NewBuffer(nil)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		p.Reset()
		p.Marshal(pb)
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	b.StopTimer()

	pb := initGoTest(true)

	// Create an array
	const N = 1000 // Internally the library starts much smaller.
	pb.F_Int32Repeated = make([]int32, N)

	// Fill in the array with some values.
	for i := 0; i < N; i++ {
		pb.F_Int32Repeated[i] = int32(i)
	}
	pbd := new(GoTest)
	p := NewBuffer(nil)
	p.Marshal(pb)
	p2 := NewBuffer(nil)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		p2.SetBuf(p.Bytes())
		p2.Unmarshal(pbd)
	}
}
