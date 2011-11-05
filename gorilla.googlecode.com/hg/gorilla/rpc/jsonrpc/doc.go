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

Usage is the same as rpc/jsonrpc. First we define a service with a method to
be called remotely:

	type ArithArgs struct {
		A, B int
	}

	type Arith int

	func (t *Arith) Multiply(r *http.Request, args ArithArgs, reply *int) os.Error {
		*reply = args.A * args.B
		return nil
	}

Then register it to be served:

	arith := new(Arith)
	s := new(jsonrpc.Server)
	s.Register(arith)

And finally setup the server as an http.Handler:

	http.Handle("/rpc", server)

In this example, one method will be registered and can be accessed as
"Arith.Multiply".

Only methods that satisfy these criteria will be made available for remote
access:

	- The method name is exported, that is, begins with an upper case letter.
	- The method receiver is exported or local (defined in the package
	  registering the service).
	- The method has three arguments.
	- The first argument is *http.Request.
	- The second and third arguments are exported or local types.
	- The third argument is a pointer.
	- The method has return type os.Error.
*/
package jsonrpc
