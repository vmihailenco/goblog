// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2011 Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package gorilla/rpc/jsonrpc is used to build web services using the
JSON-RPC 1.0 protocol.

This package is based on the standard rpc/jsonrpc package but suitable for use
on Google App Engine, as it allows http transmition over non-persistent
connections. Also, http.Request is passed as an argument to service methods,
allowing services to retrieve the App Engine context.

Usage is the same as rpc/jsonrpc. First we define a service:

	type ArithArgs struct {
		A, B int
	}

	type Arith int

	func (t *Arith) Multiply(r *http.Request, args ArithArgs, reply *int) os.Error {
		*reply = args.A * args.B
		return nil
	}

Then setup it to be served:

	arith := new(Arith)
	s := new(jsonrpc.Server)
	s.Register(arith)

And finally setup the server as an http.Handler:

	http.Handle("/rpc", server")
*/
package jsonrpc
