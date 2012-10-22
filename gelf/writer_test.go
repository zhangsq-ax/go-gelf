// Copyright 2012 SocialCode. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package gelf

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
)

func TestNewWriter(t *testing.T) {
	w, err := NewWriter("")
	if err == nil || w != nil {
		t.Errorf("New didn't fail")
		return
	}
}

func sendAndRecv(msgData string, compress CompressType) (*Message, error) {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("ResolveUDPAddr: %s", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("ListenUDP: %s", err)
	}

	w, err := NewWriter(conn.LocalAddr().String())
	if err != nil {
		return nil, fmt.Errorf("New: %s", err)
	}
	w.CompressionType = compress

	if _, err = w.Write([]byte(msgData)); err != nil {
		return nil, fmt.Errorf("w.Write: %s", err)
	}

	// the data we get from the wire is compressed
	cBuf := make([]byte, ChunkSize)

	n, err := conn.Read(cBuf)
	if err != nil {
		return nil, fmt.Errorf("Read: %s", err)
	}
	cHead, cBuf := cBuf[:2], cBuf[:n]

	var cReader io.Reader
	if bytes.Equal(cHead, magicGzip) {
		cReader, err = gzip.NewReader(bytes.NewReader(cBuf))
	} else if cHead[0] == magicZlib[0] &&
		(int(cHead[0]) * 256 + int(cHead[1])) % 31 == 0 {
		// zlib is slightly more complicated, but correct
		cReader, err = zlib.NewReader(bytes.NewReader(cBuf))
	} else {
		return nil, fmt.Errorf("unknown magic: %x", cHead)
	}

	if err != nil {
		return nil, fmt.Errorf("NewReader: %s", err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, cReader)
	if err != nil {
		return nil, fmt.Errorf("io.Copy: %s", err)
	}

	msg := new(Message)
	if err := json.Unmarshal(buf.Bytes(), &msg); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %s", err)
	}

	return msg, nil
}

// tests single-message (non-chunked) messages that are split over
// multiple lines
func TestWriteSmallMultiLine(t *testing.T) {
	for _, i := range []CompressType{CompressGzip, CompressZlib} {
		msgData := "awesomesauce\nbananas"

		msg, err := sendAndRecv(msgData, i)
		if err != nil {
			t.Errorf("sendAndRecv: %s", err)
			return
		}

		if msg.Short != "awesomesauce" {
			t.Errorf("msg.Short: expected %s, got %s", msgData, msg.Full)
			return
		}

		if msg.Full != msgData {
			t.Errorf("msg.Full: expected %s, got %s", msgData, msg.Full)
			return
		}

		fileExpected := "/go-gelf/gelf/writer_test.go"
		if !strings.HasSuffix(msg.File, fileExpected) {
			t.Errorf("msg.File: expected %s, got %s", fileExpected,
				msg.File)
		}
	}
}

// tests single-message (non-chunked) messages that are a single line long
func TestWriteSmallOneLine(t *testing.T) {
	msgData := "some awesome thing\n"
	msgDataTrunc := msgData[:len(msgData)-1]

	msg, err := sendAndRecv(msgData, CompressGzip)
	if err != nil {
		t.Errorf("sendAndRecv: %s", err)
		return
	}

	// we should remove the trailing newline
	if msg.Short != msgDataTrunc {
		t.Errorf("msg.Short: expected %s, got %s",
			msgDataTrunc, msg.Short)
		return
	}

	if msg.Full != "" {
		t.Errorf("msg.Full: expected %s, got %s", msgData, msg.Full)
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
