// Copyright 2012 SocialCode. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package gelf

import (
	"log"
	"testing"
)

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

}
