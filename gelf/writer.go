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

// Message represents the contents of the GELF message.  It is gzipped
// before sending.
//
// TODO: support optional ('_'-prefixed) fields?
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

// Used to control GELF chunking.  Should be less than (MTU - len(UDP
// header) - len(GELF header)).
//
// TODO: generate dynamically using Path MTU Discovery?
const ChunkSize = 1420

// New returns a new GELF Writer.  This writer can be used to send the
// output of the standard Go log functions to a central GELF server by
// passing it to log.SetOutput()
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

// WriteMessage sends the specified message to the GELF server
// specified in the call to New().  It assumes all the fields are
// filled out appropriately.  In general, most clients will want to
// use Write, rather than WriteMessage.
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

// getCallerIgnoringLog returns the filename and the line info of a
// function further down in the call stack.  Passing 0 in as callDepth
// would return info on the function calling getCallerIgnoringLog, 1
// the parent function, and so on.  The exception is that if the frame
// that is pointed to is from the go log library, included with go, it
// is ignored, and the function that called e.g. log.Println() is
// returned.
func getCallerIgnoringLog(callDepth int) (file string, line int) {
	// bump by 1 to ignore the getCallerIgnoringLog (this) stackframe
	callDepth++

	for {
		var ok bool
		_, file, line, ok = runtime.Caller(callDepth)
		if !ok {
			file = "???"
			line = 0
			break
		}
		if !strings.HasSuffix(file, "/pkg/log/log.go") {
			break
		}
		callDepth++
	}
	return
}

// Write encodes the given string in a GELF message and sends it to
// the server specified in New().
func (w *Writer) Write(p []byte) (n int, err error) {

	// 1 for the function that called us.
	file, line := getCallerIgnoringLog(1)

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
