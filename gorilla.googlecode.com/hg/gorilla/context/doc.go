// Copyright 2011 Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package gorilla/context provides utilities to manage request contexts.

A context stores global values for HTTP requests in a thread-safe manner.
The original idea was posted by Brad Fitzpatrick to the go-nuts mailing list:

	http://groups.google.com/group/golang-nuts/msg/e2d679d303aa5d53

Any library that needs to set request variables to be accessed by handlers
can set up a namespace to store those variables. First you create a namespace:

	var ns = new(context.Namespace)

...then call Set() or Get() as needed to define or retrieve request variables:

	// val is nil because we haven't set any value yet.
	val := ns.Get(request)

	// let's set a value, then.
	ns.Set(request, "foo")

	// val is now "foo".
	val = ns.Get(request)

You can store any type in the request context, because it accepts and returns
interface{}. To enforce a given type, wrap the getter and setter to accept and
return values of a specific type ("SomeType" in this example):

	ns = new(context.Namespace)

	// Val returns a value for this package from the request context.
	func Val(request *http.Request) SomeType {
		rv := ns.Get(request)
		if rv != nil {
			return rv.(SomeType)
		}
		return nil
	}

	// SetVal sets a value for this package in the request context.
	func SetVal(request *http.Request, val SomeType) {
		ns.Set(request, val)
	}

Notice that we now perform type casting in Val(), but set the value directly
in SetVal().

To access the namespace variable inside a handler, call the namespace
getter function passing the current request. For the previous example
we would do:

	func someHandler(w http.ResponseWriter, r *http.Request) {
		val := Val(req)

		// ...
	}

Make sure that the main handler clears the context after serving a request:

	func handler(w http.ResponseWriter, r *http.Request) {
		defer context.DefaultContext.Clear(r)

		// ...
	}

This calls Clear() from the Context instance, removing all namespaces
registered for a request.

The package gorilla/mux clears the default context, so if you are using the
default handler from there you don't need to clear anything: any namespaces
set using the default context will be cleared at the end of a request.
*/
package context
