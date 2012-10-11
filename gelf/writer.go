// Copyright 2012 SocialCode. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package gelf

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Writer struct {
	mu       sync.Mutex
	conn     net.Conn
	facility string
	hostname string
}

type Message struct {
	Version      string `json:"version"`
	Host         string `json:"host"`
	ShortMessage string `json:"short_message"`
	FullMessage  string `json:"full_message"`
	TimeUnix     int64  `json:"timestamp"`
	Level        int32  `json:"level"`
	Facility     string `json:"facility"`
	File         string `json:"file"`
	Line         int    `json:"line"`
}

/*
{
  "version": "1.0",
  "host": "www1",
  "short_message": "Short message",
  "full_message": "Backtrace here\n\nmore stuff",
  "timestamp": 1291899928.412,
  "level": 1,
  "facility": "payment-backend",
  "file": "/var/www/somefile.rb",
  "line": 356,
  "_user_id": 42,
  "_something_else": "foo"
}
*/

// Used to control GELF chunking
const ChunkSize = 1420

func New(addr string) (*Writer, error) {
	var err error
	w := new(Writer)

	if w.conn, err = net.Dial("udp", addr); err != nil {
		return nil, err
	}
	if w.hostname, err = os.Hostname(); err != nil {
		return nil, err
	}

	return w, nil
}

func (w *Writer) WriteMessage(m *Message) error {
	mBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}

	n, err := w.conn.Write(mBytes)
	if err != nil {
		return err
	}
	if n != len(mBytes) {
		return fmt.Errorf("bad write (%d/%d)", n, len(mBytes))
	}

	return nil
}

/*
func (w *Writer) Alert(m string) (err error)
func (w *Writer) Close() error
func (w *Writer) Crit(m string) (err error)
func (w *Writer) Debug(m string) (err error)
func (w *Writer) Emerg(m string) (err error)
func (w *Writer) Err(m string) (err error)
func (w *Writer) Info(m string) (err error)
func (w *Writer) Notice(m string) (err error)
func (w *Writer) Warning(m string) (err error)
*/

// getCallerIgnoringLog returns the filename and the line info
func getCallerIgnoringLog() (file string, line int) {
	var ok bool
	calldepth := 2 // 0 would be this function, 1 would be Writer.Write

	for {
		_, file, line, ok = runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
			break
		}
		if !strings.HasSuffix(file, "/pkg/log/log.go") {
			break
		}
		calldepth++
	}
	return
}

func (w *Writer) Write(p []byte) (n int, err error) {

	file, line := getCallerIgnoringLog()

	fmt.Printf("f: %s, line %d\n", file, line)

	m := Message{
		Version:      "1.0",
		Host:         w.hostname,
		ShortMessage: "",
		FullMessage:  "",
		TimeUnix:     time.Now().Unix(),
		Level:        1,
		Facility:     w.facility,
		File:         file,
		Line:         line,
	}

	if err = w.WriteMessage(&m); err != nil {
		return 0, err
	}

	return len(p), nil
}
