// Copyright 2011 Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sessions

import (
	"bytes"
	"crypto/aes"
	"crypto/hmac"
	"fmt"
	"http"
	"os"
	"testing"
	"time"
)

// ----------------------------------------------------------------------------
// ResponseRecorder
// ----------------------------------------------------------------------------
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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

// Now the tests --------------------------------------------------------------

var testSessionValues = []SessionData{
	{"foo": "bar"},
	{"baz": "ding"},
}

var testStringValues = []string{"foo", "bar", "baz", "foobar", "foobarbaz"}

func TestSerialization(t *testing.T) {
	var deserialized SessionData

	for _, value := range testSessionValues {
		serialized, err := serialize(value)
		if err != nil {
			t.Error(err)
		}
		deserialized, err = deserialize(serialized)
		if err != nil {
			t.Error(err)
		}
		if fmt.Sprintf("%v", deserialized) != fmt.Sprintf("%v", value) {
			t.Errorf("Expected %v, got %v.", value, deserialized)
		}
	}
}

func TestEncryption(t *testing.T) {
	block, e := aes.NewCipher([]byte("1234567890123456"))
	if e != nil {
		t.Error(e)
	}

	var encrypted, decrypted []byte
	var err os.Error

	for _, value := range testStringValues {
		encrypted, err = encrypt(block, []byte(value))
		decrypted, err = decrypt(block, encrypted)
		if err != nil {
			t.Error(err)
		}
		if string(decrypted) != value {
			t.Errorf("Expected %v, got %v.", value, string(decrypted))
		}
	}
}

func TestEncryptionBadBlock(t *testing.T) {
	// Invalid block size.
	block, err := aes.NewCipher([]byte("123"))
	if err == nil {
		t.Error("Expected invalid block size error.")
	}

	for _, value := range testStringValues {
		_, err = encrypt(block, []byte(value))
		if err == nil {
			t.Error("Expected invalid block size error.")
		}
	}

	_, err = decrypt(block, []byte("123456789012345612345678901234561234567890123456"))
	if err == nil {
		t.Error("Expected invalid block size error.")
	}
}

func TestAuthentication(t *testing.T) {
	// TODO test too old / too new timestamps
	hash := hmac.NewSHA256([]byte("secret-key"))
	key := "session-key"
	timestamp := time.UTC().Seconds()

	for _, value := range testStringValues {
		signed := createHmac(hash, key, []byte(value), timestamp)
		verified, err := verifyHmac(hash, key, signed, 0, 0, 0)
		if err != nil {
			t.Error(err)
		}
		if string(verified) != value {
			t.Errorf("Expected %v, got %v.", value, string(verified))
		}
	}
}

func TestEncoding(t *testing.T) {
	for _, value := range testStringValues {
		encoded := encode([]byte(value))
		decoded, err := decode(encoded)
		if err != nil {
			t.Error(err)
		}
		if string(decoded) != value {
			t.Errorf("Expected %v, got %v.", value, string(decoded))
		}
	}
}

func TestEncoder(t *testing.T) {
	b, err := aes.NewCipher([]byte("1234567890123456"))
	if err != nil {
		t.Error(err)
	}
	e1 := Encoder{
		Hash:  hmac.NewSHA256([]byte("secret-key1")),
		Block: b,
	}

	b, err = aes.NewCipher([]byte("0123456789012345"))
	if err != nil {
		t.Error(err)
	}
	e2 := Encoder{
		Hash:  hmac.NewSHA256([]byte("secret-key2")),
		Block: b,
	}

	value := make(SessionData)
	value["foo"] = "bar"
	value["baz"] = 128

	var count int
	for i := 0; i < 50; i++ {
		// Running this multiple times to check if any special character
		// breaks encoding/decoding.
		value := make(SessionData)
		value["foo"] = "bar"
		value["baz"] = 128

		encoded, err2 := e1.Encode("sid", value)
		if err2 != nil {
			t.Error(err2)
		}
		decoded, err3 := e1.Decode("sid", encoded)
		if err3 != nil {
			t.Errorf("%v: %v", err3, encoded)
			count++
		}
		if fmt.Sprintf("%v", decoded) != fmt.Sprintf("%v", value) {
			t.Errorf("Expected %v, got %v.", value, decoded)
		}
		_, err4 := e2.Decode("sid", encoded)
		if err4 == nil {
			t.Errorf("Expected failure decoding.")
		}
	}
	if count > 0 {
		t.Errorf("%d errors out of 100.", count)
	}
}

func TestLoadSaveSession(t *testing.T) {
	var req *http.Request
	var rsp *ResponseRecorder
	var hdr http.Header
	var err os.Error
	var err2 bool
	var session SessionData
	var cookies []string

	DefaultSessionFactory.SetStoreKeys("cookie",
		[]byte("my-secret-key"),
		[]byte("1234567890123456"))

	sessionValues1 := map[string]interface{}{"a": "1", "b": "2", "c": "3"}
	sessionValues2 := map[string]interface{}{"a": "11", "b": "22", "c": "33"}

	// Round 1 ----------------------------------------------------------------
	// Save an empty session.
	req, _ = http.NewRequest("GET", "http://localhost:8080/", nil)
	rsp = NewRecorder()
	if _, err = Session(req); err == nil {
		// Nothing changed, but we want to test anyway.
		Save(req, rsp)
	} else {
		t.Error(err)
	}
	hdr = rsp.Header()
	cookies, err2 = hdr["Set-Cookie"]
	if !err2 || len(cookies) != 1 {
		t.Errorf("Expected Set-Cookie key. Header: %v", hdr)
	}

	// Round 2 ----------------------------------------------------------------
	// Set some values.
	req, _ = http.NewRequest("GET", "http://localhost:8080/", nil)
	rsp = NewRecorder()
	if session, err = Session(req); err == nil {
		session["a"] = "1"
		session["b"] = "2"
		session["c"] = "3"
		Save(req, rsp)
	} else {
		t.Error(err)
	}
	hdr = rsp.Header()
	cookies, err2 = hdr["Set-Cookie"]
	if !err2 {
		t.Errorf("Expected Set-Cookie key. Header: %v", hdr)
	}

	// Round 3 ----------------------------------------------------------------
	// Change all values.
	req, _ = http.NewRequest("GET", "http://localhost:8080/", nil)
	req.Header.Add("Cookie", cookies[0])
	rsp = NewRecorder()
	if session, err = Session(req); err == nil {
		for k, v := range sessionValues1 {
			if session[k] != v {
				t.Errorf("Expected %v:%v; Got %v:%v", k, v, k, session[k])
			}
		}
		session["a"] = "11"
		session["b"] = "22"
		session["c"] = "33"
		Save(req, rsp)
	} else {
		t.Error(err)
	}
	hdr = rsp.Header()
	cookies, err2 = hdr["Set-Cookie"]
	if !err2 || len(cookies) != 1 {
		t.Errorf("Expected Set-Cookie key. Header: %v", hdr)
	}

	// Round 4 ----------------------------------------------------------------
	// Remove all values; set a new value.
	req, _ = http.NewRequest("GET", "http://localhost:8080/", nil)
	req.Header.Add("Cookie", cookies[0])
	rsp = NewRecorder()
	if session, err = Session(req); err == nil {
		for k, v := range sessionValues2 {
			if session[k] != v {
				t.Errorf("Expected %v:%v; Got %v:%v", k, v, k, session[k])
			}
		}
		session["a"] = nil, false
		session["b"] = nil, false
		session["c"] = nil, false
		session["d"] = "4"
	} else {
		t.Error(err)
	}
	// Custom key.
	if session, err = Session(req, "custom_key"); err == nil {
		session["foo"] = "bar"
	} else {
		t.Error(err)
	}
	Save(req, rsp)
	hdr = rsp.Header()
	cookies, err2 = hdr["Set-Cookie"]
	if !err2 || len(cookies) != 2 {
		t.Errorf("Expected 2 Set-Cookie key; Got %d. Header: %v", len(cookies), hdr)
	}

	// Round 5 ----------------------------------------------------------------
	// The end.
	req, _ = http.NewRequest("GET", "http://localhost:8080/", nil)
	req.Header.Add("Cookie", cookies[0])
	req.Header.Add("Cookie", cookies[1])
	rsp = NewRecorder()
	if session, err = Session(req); err == nil {
		if len(session) != 1 {
			t.Errorf("Expected a single item; Got %v", session)
		}
		if session["d"] != "4" {
			t.Errorf("Expected 4; Got %v", session["d"])
		}
		Save(req, rsp)
	} else {
		t.Error(err)
	}
	// Custom key.
	if session, err = Session(req, "custom_key"); err == nil {
		if len(session) != 1 {
			t.Errorf("Expected a single item; Got %v", session)
		}
		if session["foo"] != "bar" {
			t.Errorf("Expected bar; Got %v", session["foo"])
		}
		Save(req, rsp)
	} else {
		t.Error(err)
	}
}

func TestFlashes(t *testing.T) {
	var req *http.Request
	var rsp *ResponseRecorder
	var hdr http.Header
	var err os.Error
	var ok, err2 bool
	var cookies []string
	var flashes []interface{}

	DefaultSessionFactory.SetStoreKeys("cookie",
		[]byte("my-secret-key"),
		[]byte("1234567890123456"))

	// Round 1 ----------------------------------------------------------------
	req, _ = http.NewRequest("GET", "http://localhost:8080/", nil)
	rsp = NewRecorder()

	// Get a flash.
	if flashes, err = Flashes(req); err == nil {
		t.Errorf("Expected empty flashes; Got %v", flashes)
	}

	// Add some flashes.
	if ok, err = AddFlash(req, "foo"); !ok {
		t.Error(err)
	}
	if ok, err = AddFlash(req, "bar"); !ok {
		t.Error(err)
	}
	// Custom flash key.
	if ok, err = AddFlash(req, "baz", "custom_key"); !ok {
		t.Error(err)
	}
	// Custom flash + session key.
	if ok, err = AddFlash(req, "ding", "custom_key", "custom_session_key"); !ok {
		t.Error(err)
	}

	Save(req, rsp)
	hdr = rsp.Header()
	cookies, err2 = hdr["Set-Cookie"]
	if !err2 || len(cookies) != 2 {
		t.Errorf("Expected 2 Set-Cookie key; Got %d. Header: %v", len(cookies), hdr)
	}

	// Round 2 ----------------------------------------------------------------
	req, _ = http.NewRequest("GET", "http://localhost:8080/", nil)
	req.Header.Add("Cookie", cookies[0])
	req.Header.Add("Cookie", cookies[1])
	rsp = NewRecorder()

	// Check all saved values.
	if flashes, err = Flashes(req); err != nil || len(flashes) != 2 {
		t.Errorf("Expected flashes; Got %v", flashes)
	}
	if flashes[0] != "foo" || flashes[1] != "bar" {
		t.Errorf("Expected foo,bar; Got %v", flashes)
	}
	if flashes, err = Flashes(req); err == nil {
		t.Errorf("Expected dumped flashes; Got %v", flashes)
	}

	if flashes, err = Flashes(req, "custom_key"); err != nil || len(flashes) != 1 {
		t.Errorf("Expected flashes; Got %v", flashes)
	} else if flashes[0] != "baz" {
		t.Errorf("Expected baz; Got %v", flashes)
	}
	if flashes, err = Flashes(req, "custom_key"); err == nil {
		t.Errorf("Expected dumped flashes; Got %v", flashes)
	}

	if flashes, err = Flashes(req, "custom_key", "custom_session_key"); err != nil || len(flashes) != 1 {
		t.Errorf("Expected flashes; Got %v", flashes)
	} else if flashes[0] != "ding" {
		t.Errorf("Expected ding; Got %v", flashes)
	}
	if flashes, err = Flashes(req, "custom_key", "custom_session_key"); err == nil {
		t.Errorf("Expected dumped flashes; Got %v", flashes)
	}
}

func TestKeyRotation(t *testing.T) {
	var req *http.Request
	var rsp *ResponseRecorder
	var hdr http.Header
	var err os.Error
	var err2 bool
	var session SessionData
	var cookies []string

	DefaultSessionFactory.SetStoreKeys("cookie",
		[]byte("my-secret-key"),
		[]byte("1234567890123456"))

	// Round 1 ----------------------------------------------------------------
	// Set some values.
	req, _ = http.NewRequest("GET", "http://localhost:8080/", nil)
	rsp = NewRecorder()
	if session, err = Session(req); err == nil {
		session["a"] = "1"
		Save(req, rsp)
	} else {
		t.Error(err)
	}
	hdr = rsp.Header()
	cookies, err2 = hdr["Set-Cookie"]
	if !err2 || len(cookies) != 1 {
		t.Errorf("Expected Set-Cookie key. Header: %v", hdr)
	}

	// Round 2 ----------------------------------------------------------------
	// Invalid keys.
	DefaultSessionFactory.SetStoreKeys("cookie",
		[]byte("my-other-secret-key"),
		[]byte("1134567890123456"))

	req, _ = http.NewRequest("GET", "http://localhost:8080/", nil)
	req.Header.Add("Cookie", cookies[0])
	rsp = NewRecorder()
	if session, err = Session(req); err == nil {
		if len(session) > 0 {
			t.Errorf("Expected empty session; Got %v", session)
		}
	} else {
		t.Error(err)
	}

	// Round 3 ----------------------------------------------------------------
	// Put back the old keys.
	DefaultSessionFactory.SetStoreKeys("cookie",
		[]byte("my-other-secret-key"),
		[]byte("1134567890123456"),
		[]byte("my-secret-key"),
		[]byte("1234567890123456"),
		[]byte("my-other-secret-key"),
		[]byte("1134567890123456"))

	req, _ = http.NewRequest("GET", "http://localhost:8080/", nil)
	req.Header.Add("Cookie", cookies[0])
	rsp = NewRecorder()
	if session, err = Session(req); err == nil {
		if len(session) != 1 || session["a"] != "1" {
			t.Errorf("Expected single value session; Got %v", session)
		}
	} else {
		t.Error(err)
	}
}

// TODO test Config()
