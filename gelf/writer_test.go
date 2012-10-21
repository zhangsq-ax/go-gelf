// Copyright 2012 SocialCode. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package gelf

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"io"
	"net"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	w, err := New("")
	if err == nil || w != nil {
		t.Errorf("New didn't fail")
		return
	}
}

// tests single-message (non-chunked) messages
func TestWriteSmall(t *testing.T) {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Errorf("ResolveUDPAddr: %s", err)
		return
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Errorf("ListenUDP: %s", err)
		return
	}

	w, err := New(conn.LocalAddr().String())
	if err != nil {
		t.Errorf("New: %s", err)
		return
	}

	msgData := []byte("awesomesauce\nbananas")
	w.Write(msgData)

	// the data we get from the wire is zlib compressed
	zBuf := make([]byte, ChunkSize)

	n, err := conn.Read(zBuf)
	if err != nil {
		t.Errorf("Read: %s", err)
		return
	}
	zBuf = zBuf[:n]

	zReader, err := zlib.NewReader(bytes.NewReader(zBuf))
	if err != nil {
		t.Errorf("zlib.NewReader: %s", err)
		return
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, zReader)
	if err != nil {
		t.Errorf("io.Copy: %s", err)
		return
	}

	var msg Message
	if err := json.Unmarshal(buf.Bytes(), &msg); err != nil {
		t.Errorf("json.Unmarshal: %s", err)
		return
	}

	if !bytes.Equal(msg.Short, []byte("awesomesauce")) {
		t.Errorf("msg.Short: expected %s, got %s", string(msgData),
			string(msg.Full))
		return
	}

	if !bytes.Equal(msg.Full, msgData) {
		t.Errorf("msg.Full: expected %s, got %s", string(msgData),
			string(msg.Full))
		return
	}

	fileExpected := "/go-gelf/gelf/writer_test.go"
	if !strings.HasSuffix(msg.File, fileExpected) {
		t.Errorf("msg.File: expected %s, got %s", fileExpected,
			msg.File)
	}
}

func TestGetCaller(t *testing.T) {
	file, line := getCallerIgnoringLog(1000)
	if line != 0 || file != "???" {
		t.Errorf("didn't fail 1 %s %d", file, line)
		return
	}

	file, _ = getCallerIgnoringLog(0)
	if !strings.HasSuffix(file, "/gelf/writer_test.go") {
		t.Errorf("not writer_test.go? %s", file)
	}
}
