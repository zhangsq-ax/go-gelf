// Copyright 2012 SocialCode. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package gelf

type GELFWriter struct {

}

// Used to control GELF chunking
const ChunkSize = 1420

func NewGELFWriter(addr string) (*GELFWriter, error) {
	w := new(GELFWriter)

	return w, nil
}

func (g *GELFWriter) Write(p []byte) (n int, err error) {
	return 0, nil
}
