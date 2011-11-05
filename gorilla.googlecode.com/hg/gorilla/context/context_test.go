// Copyright 2011 Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"http"
	"testing"
)

func TestContext(t *testing.T) {
	assertEqual := func(val interface{}, exp interface{}) {
		if val != exp {
			t.Errorf("Expected %v, got %v.", exp, val)
		}
	}

	req, _ := http.NewRequest("GET", "http://localhost:8080/", nil)
	ns := new(Namespace)

	// Context.Get(), Namespace.Get(), still empty.
	assertEqual(DefaultContext.Get(req, ns), nil)
	assertEqual(ns.Get(req), nil)

	// Context.Set().
	DefaultContext.Set(req, ns, "1")
	assertEqual(DefaultContext.Get(req, ns), "1")
	assertEqual(ns.Get(req), "1")
	assertEqual(len(DefaultContext.m), 1)
	assertEqual(len(DefaultContext.m[req]), 1)

	// Context.Clear().
	DefaultContext.Clear(req)
	assertEqual(DefaultContext.Get(req, ns), nil)
	assertEqual(ns.Get(req), nil)
	assertEqual(len(DefaultContext.m), 0)

	// Context.ClearNamespace().
	DefaultContext.Set(req, ns, "2")
	assertEqual(DefaultContext.Get(req, ns), "2")
	assertEqual(ns.Get(req), "2")
	DefaultContext.ClearNamespace(req, ns)
	assertEqual(DefaultContext.Get(req, ns), nil)
	assertEqual(ns.Get(req), nil)
	assertEqual(len(DefaultContext.m), 1)
	assertEqual(len(DefaultContext.m[req]), 0)

	// Namespace.Set().
	ns.Set(req, "3")
	assertEqual(DefaultContext.Get(req, ns), "3")
	assertEqual(ns.Get(req), "3")
	assertEqual(len(DefaultContext.m), 1)
	assertEqual(len(DefaultContext.m[req]), 1)

	// Namespace.Clear().
	ns.Clear(req)
	assertEqual(DefaultContext.Get(req, ns), nil)
	assertEqual(ns.Get(req), nil)
	assertEqual(len(DefaultContext.m), 1)
	assertEqual(len(DefaultContext.m[req]), 0)
}
