// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2011 Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jsonrpc

import (
	"fmt"
	"http"
	"io/ioutil"
	"json"
	"log"
	"os"
	"strings"
	"sync"
	"unicode"
	"utf8"
	"reflect"
)

// ----------------------------------------------------------------------------
// ServiceMap
// ----------------------------------------------------------------------------

// Precompute the reflect type for os.Error.  Can't use os.Error directly
// because Typeof takes an empty interface value.  This is annoying.
var unusedError *os.Error
var typeOfOsError = reflect.TypeOf(unusedError).Elem()

// Same as above, this time for http.Request.
var unusedRequest *http.Request
var typeOfRequest = reflect.TypeOf(unusedRequest).Elem()

type methodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
}

type service struct {
	// name of service
	name   string
	// receiver of methods for the service
	rcvr   reflect.Value
	// type of the receiver
	typ    reflect.Type
	// registered methods
	method map[string]*methodType
}

// ServiceMap is a registry for services.
type ServiceMap struct {
	// protects the services map
	mu       sync.Mutex
	services map[string]*service
}

// Get returns a registered service given a method name.
//
// The method name uses a dotted notation, as in "Service.Method".
func (m *ServiceMap) Get(method string) (service *service, mtype *methodType,
err os.Error) {
	serviceMethod := strings.Split(method, ".")
	if len(serviceMethod) != 2 {
		err = os.NewError("rpc: service/method request ill-formed: " + method)
		return
	}
	m.mu.Lock()
	service = m.services[serviceMethod[0]]
	m.mu.Unlock()
	if service == nil {
		err = os.NewError("rpc: can't find service " + method)
		return
	}
	mtype = service.method[serviceMethod[1]]
	if mtype == nil {
		err = os.NewError("rpc: can't find method " + method)
		return
	}
	return
}

// Register registers a service in the services map.
func (m *ServiceMap) Register(rcvr interface{}) os.Error {
	return m.register(rcvr, "", false)
}

// RegisterName registers a named service in the services map.
func (m *ServiceMap) RegisterName(name string, rcvr interface{}) os.Error {
	return m.register(rcvr, name, true)
}

func (m *ServiceMap) register(rcvr interface{}, name string,
useName bool) os.Error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.services == nil {
		m.services = make(map[string]*service)
	}
	s := new(service)
	s.typ = reflect.TypeOf(rcvr)
	s.rcvr = reflect.ValueOf(rcvr)
	sname := reflect.Indirect(s.rcvr).Type().Name()
	if useName {
		sname = name
	}
	if sname == "" {
		log.Fatal("rpc: no service name for type", s.typ.String())
	}
	if !isExported(sname) && !useName {
		s := "rpc Register: type " + sname + " is not exported"
		log.Print(s)
		return os.NewError(s)
	}
	if _, present := m.services[sname]; present {
		return os.NewError("rpc: service already defined: " + sname)
	}
	s.name = sname
	s.method = make(map[string]*methodType)

	// Install the methods
	for m := 0; m < s.typ.NumMethod(); m++ {
		method := s.typ.Method(m)
		mtype := method.Type
		mname := method.Name
		if method.PkgPath != "" {
			continue
		}
		// Method needs four ins: receiver, *http.Request, *args, *reply.
		if mtype.NumIn() != 4 {
			log.Println("method", mname, "has wrong number of ins:",
				mtype.NumIn())
			continue
		}
		// First arg must be a pointer and must be http.Request.
		reqType := mtype.In(1)
		if reqType.Kind() != reflect.Ptr {
			log.Println(mname, "first argument type not a pointer:", reqType)
			continue
		}
		if reqType.Elem() != typeOfRequest {
			log.Println(mname, "first argument type not http.Request:",
				reqType)
			continue
		}
		// Second arg must be exported.
		argType := mtype.In(2)
		if !isExportedOrBuiltinType(argType) {
			log.Println(mname, "argument type not exported or local:", argType)
			continue
		}
		// Third arg must be a pointer and must be exported.
		replyType := mtype.In(3)
		if replyType.Kind() != reflect.Ptr {
			log.Println("method", mname, "reply type not a pointer:",
				replyType)
			continue
		}
		if !isExportedOrBuiltinType(replyType) {
			log.Println("method", mname, "reply type not exported or local:",
				replyType)
			continue
		}
		// Method needs one out: os.Error.
		if mtype.NumOut() != 1 {
			log.Println("method", mname, "has wrong number of outs:",
				mtype.NumOut())
			continue
		}
		if returnType := mtype.Out(0); returnType != typeOfOsError {
			log.Println("method", mname, "returns", returnType.String(),
				"not os.Error")
			continue
		}
		s.method[mname] = &methodType{method: method, ArgType: argType,
		ReplyType: replyType}
	}

	if len(s.method) == 0 {
		s := "rpc Register: type " + sname + " has no exported methods of suitable type"
		log.Print(s)
		return os.NewError(s)
	}
	m.services[s.name] = s
	return nil
}

// Is this an exported - upper case - name?
func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

// Is this type exported or a builtin?
func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}

// ----------------------------------------------------------------------------
// Server
// ----------------------------------------------------------------------------

// A value sent as a placeholder for the response when the server receives
// an invalid request.
type InvalidRequest struct{}
var invalidRequest = InvalidRequest{}
var null = json.RawMessage([]byte("null"))

type JsonRequest struct {
	// A String containing the name of the method to be invoked.
	Method string           `json:"method"`
	// An Array of objects to pass as arguments to the method.
	Params *json.RawMessage `json:"params"`
	// The request id. This can be of any type. It is used to match the
	// response with the request that it is replying to.
	Id     *json.RawMessage `json:"id"`
}

type JsonResponse struct {
	// The Object that was returned by the invoked method. This must be null
	// in case there was an error invoking the method.
	Result interface{}      `json:"result"`
	// An Error object if there was an error invoking the method. It must be
	// null if there was no error.
	Error  interface{}      `json:"error"`
	// This must be the same id as the request it is responding to.
	Id     *json.RawMessage `json:"id"`
}

// Server represents an RPC Server.
type Server struct {
	Map *ServiceMap
}

func (server *Server) Get(method string) (service *service, mtype *methodType,
err os.Error) {
	return server.Map.Get(method)
}

// Register publishes in the server the set of methods of the receiver value
// that satisfy the following conditions:
//
//    - exported method
//    - first argument is a pointer to the current request
//    - last two arguments are pointers to exported structs
//    - one return value, of type os.Error
func (server *Server) Register(rcvr interface{}) os.Error {
	if server.Map == nil {
		server.Map = new(ServiceMap)
	}
	return server.Map.Register(rcvr)
}

// RegisterName is like Register but uses the provided name for the type
// instead of the receiver's concrete type.
func (server *Server) RegisterName(name string, rcvr interface{}) os.Error {
	if server.Map == nil {
		server.Map = new(ServiceMap)
	}
	return server.Map.RegisterName(name, rcvr)
}

const (
	errMethod      string = "POST required, received %s."
	errContentType string = "application/json required, received %s."
	errBadJson     string = "Bad JSON: %s."
)

func requestError(w http.ResponseWriter, status int, message string,
args ...interface{}) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, message + "\n", args...)
}

// ServeHTTP implements an http.Handler that answers RPC requests.
//
//    server := rpc.NewServer()
//    server.Register(MyService)
//    http.Handle("/rpc", server)
func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. Perform basic checkings.
	if r.Method != "POST" {
		requestError(w, 405, errContentType, r.Method)
		return
	}
	if r.Header.Get("Content-Type") != "application/json" {
		requestError(w, 415, errMethod, r.Header.Get("Content-Type"))
		return
	}

	// 2. Decode request.
	body, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	request := new(JsonRequest)
	errUnmarshal := json.Unmarshal(body, request)
	if errUnmarshal != nil {
		requestError(w, 400, errBadJson, string(body))
		return
	}

	// 3. Get service to be called.
	service, mtype, errService := server.Get(request.Method)
	if errService != nil {
		requestError(w, 400, errService.String())
		return
	}

	// 4. Call the service method passing (*http.Request, *args, *reply).
	var argv reflect.Value
	argIsValue := mtype.ArgType.Kind() == reflect.Ptr
	if argIsValue {
		argv = reflect.New(mtype.ArgType.Elem())
	} else {
		argv = reflect.New(mtype.ArgType)
	}
	arg := argv.Interface()
	if arg != nil {
		// JSON params is array value.
		// RPC params is struct.
		// Unmarshal into array containing struct for now.
		// Should think about making RPC more general.
		var params [1]interface{}
		params[0] = arg
		errUnmarshal = json.Unmarshal(*request.Params, &params)
		if errUnmarshal != nil {
			requestError(w, 400, errBadJson, arg)
			return
		}
	}
	if !argIsValue {
		argv = argv.Elem()
	}
	replyv := reflect.New(mtype.ReplyType.Elem())
	result := mtype.method.Func.Call([]reflect.Value{service.rcvr,
									 reflect.ValueOf(r), argv, replyv})

	// 5. Encode and set response.
	response := new(JsonResponse)
	// Set the result and error object, if any.
	errResult := result[0].Interface()
	if errResult != nil {
		response.Result = invalidRequest
		response.Error = errResult
		w.WriteHeader(400)
	} else {
		response.Result = replyv.Interface()
		response.Error = nil
	}
	// Set the Id: same as the one from request or null for notifications.
	if request.Id != nil {
		response.Id = request.Id
	} else {
		response.Id = &null
	}
	// Set the response.
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if request.Id != nil {
		// Don't set the response body for notifications.
		encoder := json.NewEncoder(w)
		encoder.Encode(response)
	}
}
