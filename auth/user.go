package auth

import (
	"os"

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

func GetUserByUserId(c appengine.Context, userId string) (*User, os.Error) {
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
		initUser(&users[0])
		return &users[0], nil
	}
	return nil, os.NewError("got multiple results for unique field")
}

func CreateUserFromAppengine(c appengine.Context, appengineUser *user.User) (*User, os.Error) {
	u := &User{
		UserId: appengineUser.Id,

		Name:       appengineUser.String(),
		Email:      appengineUser.Email,
		AuthDomain: appengineUser.AuthDomain,

		FederatedIdentity: appengineUser.FederatedIdentity,
		FederatedProvider: appengineUser.FederatedProvider,
	}
	initUser(u)
	if _, err := entity.PutEntity(c, u); err != nil {
		return nil, err
	}
	return u, nil
}

func CurrentUser(c appengine.Context) (*User, os.Error) {
	appengineUser := user.Current(c)
	if appengineUser == nil {
		return Anonymous, nil
	}
	u, err := GetUserByUserId(c, appengineUser.Id)
	if err != nil {
		return nil, err
	}
	if u != nil {
		return u, nil
	}
	return CreateUserFromAppengine(c, appengineUser)
}

type User struct {
	*entity.Entity `datastore:"-"`
	kind           string

	UserId string

	Name       string
	Email      string
	AuthDomain string

	FederatedIdentity string
	FederatedProvider string
}

func (u *User) SetKey(key *datastore.Key) os.Error {
	if u.Entity == nil {
		u.Entity = entity.NewEntity(USER_KIND)
	}
	return u.Entity.SetKey(key)
}

func (u *User) IsAuthenticated() bool {
	return u != Anonymous
}

func (u *User) IsAnonymous() bool {
	return !u.IsAuthenticated()
}

var Anonymous = &User{}
