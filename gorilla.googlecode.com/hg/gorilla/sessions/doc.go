// Copyright 2011 Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package gorilla/sessions provides cookie sessions and infrastructure for
custom session back-ends.

The key features are:

* Dead simple basic API: use it as an easy way to set signed (and optionally
encrypted) cookies.

* Advanced API for custom back-ends: built-in support for custom storage
systems; session store interface and helper functions; encoder interface
and default implementation with customizable cryptography methods (thanks to
Go interfaces).

* Conveniences: flash messages (session values that last until read);
built-in mechanism to rotate authentication and encryption keys;
multiple sessions per request, even using different back-ends; easy way to
switch session persistency (aka "remember me") and set other attributes.

The most basic example to retrieve a session is to call sessions.Session()
passing the current request. For example, in a handler:

	func MyHandler(w http.ResponseWriter, r *http.Request) {
		if session, err := sessions.Session(r); err == nil {
			session["foo"] = "bar"
			session["baz"] = 128
			sessions.Save(r, w)
		}
	}

The above snippet is "gorilla/sessions in a nutshell": a session is a simple
map[string]interface{} returned from sessions.Session(). It stores any values
that can be encoded using gob. After we set some values, we call
sessions.Save() passing the current request and response.

Side note about "any values that can be encoded using gob": to store special
structures in a session, we must register them using gob.Register() first.
For basic types this is not needed; it works out of the box.

Is it that simple? Well, almost. Before we can use sessions, we must define
a secret key to be used for authentication, and optionally an encryption key.
They are both set calling SetStoreKeys() and should be done at initialization
time:

	func init() {
		sessions.SetStoreKeys("cookie",
							  []byte("my-hmac-key"),
							  []byte("my-aes-key"))
	}

The first argument is the name used to register the session store. By default
a "cookie" store is registered and available for use, so we use that name.

The second argument is the secret key used to authenticate the session cookie
using HMAC. It is required; if no authentication key is set sessions can't be
read or written (and a call to sessions.Session() will return an error).

The third argument is the encryption key; it is optional and can be omitted.
For the cookie store, setting this will encrypt the contents stored in the
cookie; otherwise the contents can be read, although not forged.

Side note about the encryption key: if set, must be either 16, 24, or 32 bytes
to select AES-128, AES-192, or AES-256 modes. Otherwise a block can't be
created and sessions won't work.

Exposing the contents of a session is not a big deal in many cases, like when
we store a simple username or user id, but to to store sensitive information
using the cookie store, we must set an encryption key. For custom stores that
only set a random session id in the cookie, encryption is not needed.

And this is all you need to know about the basic usage. More advanced options
are explained below.

Sometimes we may want to change authentication and/or encryption keys without
breaking existing sessions. We can do this setting multiple authentication and
encryption keys, in pairs, to be tested in order:

	sessions.SetStoreKeys("cookie",
						  []byte("my-hmac-key"),
						  []byte("my-aes-key"),
						  []byte("my-previous-hmac-key"),
						  []byte("my-previous-aes-key"))

New sessions will be saved using the first pair. Old sessions can still be
read because the first pair will fail, and the second will be tested. This
makes it easy to "rotate" secret keys and still be able to validate existing
sessions. Note: for all pairs the encryption key is optional; set it
to nil and encryption won't be used.

Back to how sessions are retrieved.

Sessions are named. When we get a session calling sessions.Session(request),
we are implicitly requesting a session using the default key ("s") and store
(the CookieSessionStore). This is just a convenience; we can have as many
sessions as needed, just passing different session keys. For example:

	if authSession, err := sessions.Session(r, "auth"); err == nil {
		userId = authSession["userId"]
		// ...
	}

Here we requested a session explicitly naming it "auth". It will be saved
separately. This can be used as a convenient way to save signed cookies, and
is also how we access multiple sessions per request.

Session stores also have a name, and need to be registered to be available.
The default session store uses authenticated (and optionally encrypted)
cookies, and is registered with the name "cookie". To use a custom
session back-end, we first register it in the SessionFactory, then pass its
name as the third argument to sessions.Session().

For the sake of demonstration, let's pretend that we defined a store called
MemcacheSessionStore. First we register it using the "memcache" key. This
should be done at initialization time:

	func init() {
		sessions.SetStore("memcache", new(MemcacheSessionStore))
	}

Then to get a session using memcached, we pass a third argument to
sessions.Session(), the store key:

	session, err := sessions.Session(r, "mysession", "memcache")

...and it will use the custom back-end we defined, instead of the default
"cookie" one. This means that we can use multiple sessions in the same
request even using different back-ends.

And how to configure session expiration time, path or other cookie attributes?

By default, session cookies last for a month. This is probably too long for a
lot of cases, but it is easy to change this and other attributes during
runtime. Just request the session configuration struct and change the variables
as needed. The fields are basically a subset of http.Cookie fields.
To change MaxAge, we would do:

	if config, err = sessions.Config(r); err == nil {
		// Change max-age to 1 week.
		config.MaxAge = 86400 * 7
	}

After this, cookies will last for a week only. The Config() function
accepts an optional argument besides the request: the session key. If not
defined, the configuration for the default session key is returned.

Bonus: flash messages. What are they? It basically means "session values that
last until read". The term was popularized by Ruby On Rails a few years back.
When we request a flash message, it is removed from the session. We have two
convenience functions to read and set them: Flashes() and AddFlash(). Here is
an example:

	func MyHandler(w http.ResponseWriter, r *http.Request) {
		// Get the previously set flashes, if any.
		if flashes, _ := sessions.Flashes(r); flashes != nil {
			// Just print the flash values.
			fmt.Fprint(w, "%v", flashes)
		} else {
			fmt.Fprint(w, "No flashes found.")
			// Set a new flash.
			sessions.AddFlash(r, "Hello, flash messages world!")
		}
		sessions.Save(r, w)
	}

Flash messages are useful to set information to be read after a redirection,
usually after form submissions.
*/
package sessions
