package cookie

import (
	// "github.com/satori/go.uuid"
	"fmt"
	r "gopkg.in/dancannon/gorethink.v2"
	"time"
)

const (
	tablename = "remembermecookie"
)

type RethinkDBStore struct {
	session *r.Session
	db      string
}

type data struct {
	Selector   string    `gorethink:"id,omitempty"`
	User       string    `gorethink:"user,omitempty"`
	Hash       string    `gorethink:"hash"`
	Expiration time.Time `gorethink:"expiration"`
}

func NewRethinkDBStore(address, database string) (store *RethinkDBStore, err error) {
	store = new(RethinkDBStore)
	store.session, err = r.Connect(r.ConnectOpts{
		Address:  address,
		Database: database,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	store.db = database

	// Discard error (database exists)
	_, _ = r.DBCreate(database).RunWrite(store.session)
	_, _ = r.DB(database).TableCreate(tablename).RunWrite(store.session)

	return
}

func (s *RethinkDBStore) Insert(user, hash string, expiration time.Time) (selector string, err error) {
	d := data{
		User:       user,
		Hash:       hash,
		Expiration: expiration,
	}

	result, err := r.Table(tablename).Insert(d).RunWrite(s.session)

	selector = result.GeneratedKeys[0]
	return
}

func (s *RethinkDBStore) Update(selector, user, hash string, expiration time.Time) (err error) {
	d := data{
		Hash:       hash,
		Expiration: expiration,
	}

	_, err = r.Table(tablename).Get(selector).Update(d).RunWrite(s.session)
	if err != nil {
		fmt.Println(err)
		return
	}

	return
}

func (s *RethinkDBStore) Get(selector string) (user, hash string, expiration time.Time, err error) {
	cursor, err := r.Table(tablename).Get(selector).Run(s.session)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cursor.Close()

	var d data
	err = cursor.One(&d)
	if err != nil {
		return
	}

	user = d.User
	hash = d.Hash
	expiration = d.Expiration
	return
}

func (s *RethinkDBStore) DeleteSelector(selector string) (err error) {
	_, err = r.Table(tablename).Get(selector).Delete().Run(s.session)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func (s *RethinkDBStore) DeleteUser(user string) (err error) {
	_, err = r.Table(tablename).Filter(map[string]interface{}{
		"user": user,
	}).Delete().Run(s.session)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func (s *RethinkDBStore) DeleteAll() (err error) {
	_, err = r.Table(tablename).Delete().Run(s.session)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func (s *RethinkDBStore) Close() error {
	return s.session.Close()
}
