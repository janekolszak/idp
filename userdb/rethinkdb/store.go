package rethinkdb

import (
	r "gopkg.in/dancannon/gorethink.v2"

	"github.com/janekolszak/idp/core"
)

type Store struct {
	session *r.Session
}

const (
	table = "users"
)

func NewStore(session *r.Session) (*Store, error) {
	store := new(Store)
	store.session = session

	// Discard error (database exists)
	db := session.Database()
	r.DBCreate(db).RunWrite(session)
	r.DB(db).TableCreate(table).RunWrite(session)

	// Index for fetching users by credentials used in the login
	r.Table(table).IndexCreate("username").Exec(session)
	r.Table(table).IndexWait().RunWrite(session)

	return store, nil
}

func (s *Store) Get(username string) (user *User, err error) {
	result, err := r.Table(table).GetAllByIndex("username", username).Run(s.session)
	if err != nil {
		return
	}
	defer result.Close()

	if result.IsNil() {
		err = core.ErrorNoSuchUser
		return
	}

	user = new(User)
	return
}

func (s *Store) Check(username, password string) error {
	return nil
}

func (s *Store) Insert(user *User) (id string, err error) {
	result, err := r.Table(table).Insert(user).RunWrite(s.session)
	if err != nil {
		return
	}

	// TODO: Test
	// if result.IsNil() {
	// 	err = core.ErrorNoSuchUser
	// 	return
	// }

	id = result.GeneratedKeys[0]
	return
}

func (s *Store) Update(user *User) error {
	return r.Table(table).Get(user.ID).Update(user).Exec(s.session)
}

func (s *Store) Delete(id string) error {
	return r.Table(table).Get(id).Delete().Exec(s.session)
}

func (s *Store) SetIsVerified(id string) error {
	var data = map[string]interface{}{
		"isVerified": true,
	}
	return r.Table(table).Get(id).Update(data).Exec(s.session)
}
