// Copyright 2012 SocialCode. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package gelf

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"fmt"
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

func sendAndRecv(msgData []byte) (*Message, error) {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("ResolveUDPAddr: %s", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("ListenUDP: %s", err)
	}

	w, err := New(conn.LocalAddr().String())
	if err != nil {
		return nil, fmt.Errorf("New: %s", err)
	}

	w.Write(msgData)

	// the data we get from the wire is zlib compressed
	zBuf := make([]byte, ChunkSize)

	n, err := conn.Read(zBuf)
	if err != nil {
		return nil, fmt.Errorf("Read: %s", err)
	}
	zBuf = zBuf[:n]

	zReader, err := zlib.NewReader(bytes.NewReader(zBuf))
	if err != nil {
		return nil, fmt.Errorf("zlib.NewReader: %s", err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, zReader)
	if err != nil {
		return nil, fmt.Errorf("io.Copy: %s", err)
	}

	msg := new(Message)
	if err := json.Unmarshal(buf.Bytes(), &msg); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %s", err)
	}

	return msg, nil
}

// tests single-message (non-chunked) messages
func TestWriteSmall(t *testing.T) {
	msgData := []byte("awesomesauce\nbananas")

	msg, err := sendAndRecv(msgData)
	if err != nil {
		t.Errorf("sendAndRecv: %s", err)
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
