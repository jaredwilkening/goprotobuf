# Go support for Protocol Buffers - Google's data interchange format
#
# Copyright 2010 Google Inc.  All rights reserved.
# http://code.google.com/p/goprotobuf/
#
# Redistribution and use in source and binary forms, with or without
# modification, are permitted provided that the following conditions are
# met:
#
#     * Redistributions of source code must retain the above copyright
# notice, this list of conditions and the following disclaimer.
#     * Redistributions in binary form must reproduce the above
# copyright notice, this list of conditions and the following disclaimer
# in the documentation and/or other materials provided with the
# distribution.
#     * Neither the name of Google Inc. nor the names of its
# contributors may be used to endorse or promote products derived from
# this software without specific prior written permission.
#
# THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
# "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
# LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
# A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
# OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
# SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
# LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
# DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
# THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
# (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
# OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

include $(GOROOT)/src/Make.inc

TARG=code.google.com/p/goprotobuf/compiler/plugin
GOFILES=\
	plugin.pb.go\

DEPS=../../proto ../descriptor

include $(GOROOT)/src/Make.pkg

# Not stored here, but plugin.proto is in http://code.google.com/p/protobuf
# at protobuf-2.3.0/src/google/protobuf/compiler/plugin.proto
# Also we need to fix an import.
regenerate:
	echo WARNING! THIS RULE IS PROBABLY NOT RIGHT FOR YOUR INSTALLATION
	cd $(HOME)/protobuf-2.3.0/src && \
	protoc --go_out=. ./google/protobuf/compiler/plugin.proto && \
	cat ./google/protobuf/compiler/plugin.pb.go | \
		sed '/^import/s;google/protobuf/descriptor.pb;code.google.com/p/goprotobuf/compiler/descriptor;' >$(GOROOT)/src/pkg/code.google.com/p/goprotobuf/compiler/plugin/plugin.pb.go

restore:
	cp plugin.pb.golden plugin.pb.go

preserve:
	cp plugin.pb.go plugin.pb.golden
