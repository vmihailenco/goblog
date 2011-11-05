// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2011 Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jsonrpc

import (
	"bytes"
	"http"
	"json"
	"os"
	"strings"
	"testing"
)

// ResponseRecorder is an implementation of http.ResponseWriter that
// records its mutations for later inspection in tests.
type ResponseRecorder struct {
	Code      int           // the HTTP response code from WriteHeader
	HeaderMap http.Header   // the HTTP response headers
	Body      *bytes.Buffer // if non-nil, the bytes.Buffer to append written data to
	Flushed   bool
}

// NewRecorder returns an initialized ResponseRecorder.
func NewRecorder() *ResponseRecorder {
	return &ResponseRecorder{
		HeaderMap: make(http.Header),
		Body:      new(bytes.Buffer),
	}
}

// DefaultRemoteAddr is the default remote address to return in RemoteAddr if
// an explicit DefaultRemoteAddr isn't set on ResponseRecorder.
const DefaultRemoteAddr = "1.2.3.4"

// Header returns the response headers.
func (rw *ResponseRecorder) Header() http.Header {
	return rw.HeaderMap
}

// Write always succeeds and writes to rw.Body, if not nil.
func (rw *ResponseRecorder) Write(buf []byte) (int, os.Error) {
	if rw.Body != nil {
		rw.Body.Write(buf)
	}
	if rw.Code == 0 {
		rw.Code = http.StatusOK
	}
	return len(buf), nil
}

// WriteHeader sets rw.Code.
func (rw *ResponseRecorder) WriteHeader(code int) {
	rw.Code = code
}

// Flush sets rw.Flushed to true.
func (rw *ResponseRecorder) Flush() {
	rw.Flushed = true
}

// ----------------------------------------------------------------------------

type ArithArgs struct {
	A, B int
}

type Arith int

func (t *Arith) Multiply(r *http.Request, args ArithArgs, reply *int) os.Error {
	*reply = args.A * args.B
	return nil
}

func TestRegister(t *testing.T) {
	arith := new(Arith)

	s := new(Server)
	s.Register(arith)
	_, _, err := s.Map.Get("Arith.Multiply")
	if err != nil {
		t.Errorf("Expected to be registered: Arith.Multiply")
	}

	s = new(Server)
	s.RegisterName("Foo", arith)
	_, _, err = s.Map.Get("Foo.Multiply")
	if err != nil {
		t.Errorf("Expected to be registered: Foo.Multiply")
	}
}

func TestServe(t *testing.T) {
	arith := new(Arith)
	s := new(Server)
	s.Register(arith)

	w := NewRecorder()
	body := strings.NewReader(`{"method":"Arith.Multiply", "id":"anything", "params":[{"A":4, "B":2}]}`)
	r, _ := http.NewRequest("POST", "http://localhost:8080/", body)
	r.Header.Set("Content-Type", "application/json")
	s.ServeHTTP(w, r)

	response := new(JsonResponse)
	decoder := json.NewDecoder(w.Body)
	decoder.Decode(response)
	if response.Result != float64(8) {
		t.Errorf("Wrong response: %v.", response.Result)
	}
}
