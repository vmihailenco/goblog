package core

import (
	"net/http"

	"appengine"

	"auth"
)

func AuthUser(c appengine.Context, w http.ResponseWriter) *auth.User {
	user := auth.CurrentUser(c)

	if !user.IsAuth() {
		HandleAuthRequired(c, w)
		return nil
	}

	return user
}

func AdminUser(c appengine.Context, w http.ResponseWriter) *auth.User {
	user := AuthUser(c, w)
	if user == nil {
		return nil
	}

	if !user.IsAdmin {
		HandleAuthRequired(c, w)
		return nil
	}

	return user
}
