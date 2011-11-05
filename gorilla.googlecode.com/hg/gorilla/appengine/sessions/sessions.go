// Copyright 2011 Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sessions

import (
	"fmt"
	"http"
	"os"
	"time"
	"appengine"
	"appengine/datastore"
	"appengine/memcache"
	"gorilla.googlecode.com/hg/gorilla/sessions"
)

type baseSessionStore struct {
	// List of encoders registered for this store.
	encoders []sessions.SessionEncoder
}

// Encoders returns the encoders for this store.
func (s *baseSessionStore) Encoders() []sessions.SessionEncoder {
	return s.encoders
}

// SetEncoders sets a group of encoders in the store.
func (s *baseSessionStore) SetEncoders(encoders ...sessions.SessionEncoder) {
	s.encoders = encoders
}

// ----------------------------------------------------------------------------
// DatastoreSessionStore
// ----------------------------------------------------------------------------

type Session struct {
	Date  datastore.Time
	Value []byte
}

// DatastoreSessionStore stores session data in App Engine's datastore.
type DatastoreSessionStore struct {
	baseSessionStore
}

// Load loads a session for the given key.
func (s *DatastoreSessionStore) Load(r *http.Request, key string,
info *sessions.SessionInfo) {
	data := sessions.GetCookie(s, r, key)
	if sidval, ok := data["sid"]; ok {
		// Cleanup session data.
		sid := sidval.(string)
		for k, _ := range data {
			data[k] = nil, false
		}
		// Get session from datastore and deserialize it.
		c := appengine.NewContext(r)
		var session Session
		key := datastore.NewKey(c, "Session", sessionKey(sid), 0, nil)
		if err := datastore.Get(c, key, &session); err == nil {
			data, _ = sessions.DeserializeSessionData(session.Value)
		}
	}
	info.Data = data
}

// Save saves the session in the response.
func (s *DatastoreSessionStore) Save(r *http.Request, w http.ResponseWriter,
key string, info *sessions.SessionInfo) (flag bool, err os.Error) {
	sid, serialized, error := getIdAndData(info)
	if error != nil {
		err = error
		return
	}

	// Save the session.
	c := appengine.NewContext(r)
	entityKey := datastore.NewKey(c, "Session", sessionKey(sid), 0, nil)
	_, err = datastore.Put(appengine.NewContext(r), entityKey, &Session{
		Date:  datastore.SecondsToTime(time.Seconds()),
		Value: serialized,
	})
	if err != nil {
		return
	}

	return sessions.SetCookie(s, w, key, cloneInfo(info, sid))
}

// ----------------------------------------------------------------------------
// MemcacheSessionStore
// ----------------------------------------------------------------------------

// MemcacheSessionStore stores session data in App Engine's memcache.
type MemcacheSessionStore struct {
	baseSessionStore
}

// Load loads a session for the given key.
func (s *MemcacheSessionStore) Load(r *http.Request, key string,
info *sessions.SessionInfo) {
	data := sessions.GetCookie(s, r, key)
	if sidval, ok := data["sid"]; ok {
		// Cleanup session data.
		sid := sidval.(string)
		for k, _ := range data {
			data[k] = nil, false
		}
		// Get session from memcache and deserialize it.
		c := appengine.NewContext(r)
		if item, err := memcache.Get(c, sessionKey(sid)); err == nil {
			data, _ = sessions.DeserializeSessionData(item.Value)
		}
	}
	info.Data = data
}

// Save saves the session in the response.
func (s *MemcacheSessionStore) Save(r *http.Request, w http.ResponseWriter,
key string, info *sessions.SessionInfo) (flag bool, err os.Error) {
	sid, serialized, error := getIdAndData(info)
	if error != nil {
		err = error
		return
	}

	// Add the item to the memcache, if the key does not already exist.
	err = memcache.Add(appengine.NewContext(r), &memcache.Item{
		Key:   sessionKey(sid),
		Value: serialized,
	})
	if err != nil {
		return
	}

	return sessions.SetCookie(s, w, key, cloneInfo(info, sid))
}

func sessionKey(sid string) string {
	return fmt.Sprintf("gorilla.appengine.sessions.%s", sid)
}

// Create a new sid and serialize data.
func getIdAndData(info *sessions.SessionInfo) (sid string, serialized []byte, err os.Error) {
	// Create a new session id.
	sid, err = sessions.GenerateSessionId(128)
	if err != nil {
		return
	}
	// Serialize session into []byte.
	serialized, err = sessions.SerializeSessionData(info.Data)
	if err != nil {
		return
	}
	return
}

// Clone info, setting only sid as data.
func cloneInfo(info *sessions.SessionInfo, sid string) *sessions.SessionInfo {
	return &sessions.SessionInfo{
		Data:   sessions.SessionData{"sid": sid},
		Store:  info.Store,
		Config: info.Config,
	}
}
