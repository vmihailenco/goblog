// Copyright 2011 Rodrigo Moraes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package gorilla/appengine/sessions implements session stores for Google App
Engine's datastore and memcache.

Usage is the same as described in gorilla/sessions documentation:

	http://gorilla-web.appspot.com/pkg/gorilla/sessions

...with a little preparation needed: the new stores must be registered and
their secret keys must be set before use. We can do it in a init() function:

	import (
		// ...
		appengineSessions "gorilla.googlecode.com/hg/gorilla/appengine/sessions"
		"gorilla.googlecode.com/hg/gorilla/sessions"
	)

	func init() {
		// Register the datastore and memcache session stores.
		sessions.SetStore("datastore", new(appengineSessions.DatastoreSessionStore))
		sessions.SetStore("memcache", new(appengineSessions.MemcacheSessionStore))

		// Set secret keys for the session stores.
		sessions.SetStoreKeys("datastore", []byte("my-secret-key"))
		sessions.SetStoreKeys("memcache", []byte("my-secret-key"))
	}

After this, to retrieve a session using datastore or memcache, pass a third
parameter to sessions.Session() with the key used to register the store.
For datastore:

	if session, err := sessions.Session(r, "", "datastore"); err == nil {
		session["foo"] = "bar"
		sessions.Save(r, w)
	}

Or for memcache:

	if session, err := sessions.Session(r, "", "memcache"); err == nil {
		session["foo"] = "bar"
		sessions.Save(r, w)
	}
*/
package sessions
