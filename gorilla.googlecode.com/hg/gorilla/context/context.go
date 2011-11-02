// Copyright 2011 Rodrigo Moraes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"http"
	"sync"
)

// DefaultContext is the default context instance.
var DefaultContext = new(Context)

// Maps request -> namespace -> value.
type contextMap map[*http.Request]namespaceMap

// Maps namespace -> value.
type namespaceMap map[Namespacer]interface{}

// ----------------------------------------------------------------------------
// Context
// ----------------------------------------------------------------------------
// Original implementation by Brad Fitzpatrick:
// http://groups.google.com/group/golang-nuts/msg/e2d679d303aa5d53

// Context stores values for requests.
type Context struct {
	lk sync.Mutex
	m  contextMap
}

// Get returns the value for a given namespace in a given request.
func (c *Context) Get(req *http.Request, ns Namespacer) interface{} {
	c.lk.Lock()
	defer c.lk.Unlock()
	if c.m != nil {
		if c.m[req] != nil {
			return c.m[req][ns]
		}
	}
	return nil
}

// Set stores a value for a given namespace in a given request.
func (c *Context) Set(req *http.Request, ns Namespacer, val interface{}) {
	c.lk.Lock()
	defer c.lk.Unlock()
	if c.m == nil {
		c.m = make(contextMap)
	}
	if c.m[req] == nil {
		c.m[req] = make(namespaceMap)
	}
	c.m[req][ns] = val
}

// Clear removes all namespaces for a given request.
func (c *Context) Clear(req *http.Request) {
	c.lk.Lock()
	defer c.lk.Unlock()
	if c.m != nil {
		if c.m[req] != nil {
			c.m[req] = nil, false
		}
	}
}

// ClearNamespace removes the value from a given namespace in a given request.
func (c *Context) ClearNamespace(req *http.Request, ns Namespacer) {
	c.lk.Lock()
	defer c.lk.Unlock()
	if c.m != nil {
		if c.m[req] != nil {
			c.m[req][ns] = nil, false
		}
	}
}

// ----------------------------------------------------------------------------
// Namespace
// ----------------------------------------------------------------------------

// Namespacer is the interface for context namespaces.
type Namespacer interface {
	Set(*http.Request, interface{})
	Get(*http.Request) interface{}
	Clear(*http.Request)
}

// Namespace is a namespace to store a value in the request context.
//
// Packages can use one or more namespaces to attach variables to the request.
// Fist you define a namespace:
//
//     var ns = new(context.Namespace)
//
// ...then call Set() or Get() as needed to define or retrieve request variables:
//
//     // val is nil because we haven't set any value yet.
//     val := ns.Get(request)
//
//     // let's set a value, then.
//     ns.Set(request, "foo")
//
//     // val is now "foo".
//     val = ns.Get(request)
//
// Each namespace can store a single value, but it can be of any type since
// Context accepts and returns a interface{} type.
type Namespace struct {
	Context *Context
}

// GetContext returns the request context this namespace is attached to.
//
// If no context was explicitly defined, it will use DefaultContext.
func (n *Namespace) GetContext() *Context {
	if n.Context == nil {
		n.Context = DefaultContext
	}
	return n.Context
}

// Get returns the value stored for this namespace in the request context.
func (n *Namespace) Get(request *http.Request) interface{} {
	return n.GetContext().Get(request, n)
}

// Set stores a value for this namespace in the request context.
func (n *Namespace) Set(request *http.Request, val interface{}) {
	n.GetContext().Set(request, n, val)
}

// Clear removes this namespace from the request context.
func (n *Namespace) Clear(request *http.Request) {
	n.GetContext().ClearNamespace(request, n)
}
