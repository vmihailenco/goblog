package auth

import (
	"errors"

	"appengine"
	"appengine/datastore"
	"appengine/user"

	"core/entity"
)

const (
	USER_KIND = "user"
)

func GetUserQuery() *datastore.Query {
	return datastore.NewQuery(USER_KIND)
}

func NewUser() *User {
	u := &User{}
	initUser(u)
	return u
}

func initUser(user *User) {
	user.Entity = entity.NewEntity(USER_KIND)
}

func GetUserByUserId(c appengine.Context, userId string) (*User, error) {
	q := GetUserQuery().Filter("UserId =", userId).Limit(2)
	users := make([]User, 0, 2)
	keys, err := q.GetAll(c, &users)
	if err != nil {
		return nil, err
	}
	switch len(keys) {
	case 0:
		return nil, nil
	case 1:
		u := &users[0]
		initUser(u)
		u.SetKey(keys[0])
		return u, nil
	}
	return nil, errors.New("Got multiple results for unique field.")
}

func CreateUserFromAppengine(c appengine.Context, appengineUser *user.User) (*User, error) {
	u := &User{
		UserId: appengineUser.ID,

		Name:       appengineUser.String(),
		Email:      appengineUser.Email,
		AuthDomain: appengineUser.AuthDomain,
		IsAdmin:    user.IsAdmin(c),

		FederatedIdentity: appengineUser.FederatedIdentity,
		FederatedProvider: appengineUser.FederatedProvider,
	}
	initUser(u)
	if err := entity.Put(c, u); err != nil {
		return nil, err
	}
	return u, nil
}

func CurrentUser(c appengine.Context) *User {
	appengineUser := user.Current(c)
	if appengineUser == nil {
		return Anonymous
	}

	u, err := GetUserByUserId(c, appengineUser.ID)
	if err != nil {
		return Anonymous
	}

	if u == nil {
		u, err = CreateUserFromAppengine(c, appengineUser)
		if err != nil {
			return Anonymous
		}
	}

	if user.IsAdmin(c) && !u.IsAdmin {
		u.IsAdmin = true
		// ignore error
		entity.Put(c, u)
	}

	return u
}

type User struct {
	*entity.Entity `datastore:"-"`
	kind           string

	UserId string

	Name       string
	Email      string
	AuthDomain string
	IsAdmin    bool

	FederatedIdentity string
	FederatedProvider string
}

func (u *User) SetKey(key *datastore.Key) {
	if u.Entity == nil {
		u.Entity = entity.NewEntity(USER_KIND)
	}
	u.Entity.SetKey(key)
}

func (u *User) IsAuth() bool {
	return u != Anonymous
}

func (u *User) IsAnonymous() bool {
	return !u.IsAuth()
}

var Anonymous = &User{}
