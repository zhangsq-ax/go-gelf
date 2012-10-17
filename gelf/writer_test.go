// Copyright 2012 SocialCode. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package gelf

import (
	"log"
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

func TestWrite(t *testing.T) {
	w, err := New("localhost:1234")
	if err != nil {
		t.Errorf("New: %s", err)
		return
	}

	log.SetOutput(w)
	log.Println("ok")

	_ = w
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
