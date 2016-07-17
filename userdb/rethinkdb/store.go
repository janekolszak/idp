package rethinkdb

import (
	"github.com/janekolszak/idp/core"

	"golang.org/x/crypto/bcrypt"
	r "gopkg.in/dancannon/gorethink.v2"
	"time"
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
	r.Table(table).IndexCreate("email").Exec(session)

	r.Table(table).IndexWait().RunWrite(session)

	return store, nil
}

func (s *Store) GetWithID(id string) (user *User, err error) {
	cursor, err := r.Table(table).Get(id).Run(s.session)
	if err != nil {
		return
	}
	defer cursor.Close()

	if cursor.IsNil() {
		err = core.ErrorNoSuchUser
		return
	}

	user = new(User)
	err = cursor.One(user)
	return
}

func (s *Store) GetWithUsername(username string) (user *User, err error) {
	cursor, err := r.Table(table).GetAllByIndex("username", username).Run(s.session)
	if err != nil {
		return
	}
	defer cursor.Close()

	if cursor.IsNil() {
		err = core.ErrorNoSuchUser
		return
	}

	user = new(User)
	err = cursor.One(user)
	return
}

func (s *Store) Check(username, password string) error {
	cursor, err := r.Table(table).GetAllByIndex("username", username).Pluck("password").Run(s.session)
	if err != nil {
		bcrypt.CompareHashAndPassword([]byte(""), []byte(password))
		return err
	}
	defer cursor.Close()

	var data map[string][]byte
	err = cursor.One(&data)
	if err != nil {
		// No such user, prevent timing atack
		bcrypt.CompareHashAndPassword([]byte(""), []byte(password))
		return core.ErrorAuthenticationFailure
	}

	err = bcrypt.CompareHashAndPassword(data["password"], []byte(password))
	if err != nil {
		return core.ErrorAuthenticationFailure
	}

	return nil
}

func (s *Store) count(indexName, value string) (uint, error) {
	cursor, err := r.Table(table).GetAllByIndex(indexName, value).Count().Run(s.session)
	if err != nil {
		return 0, err
	}

	var result interface{}
	err = cursor.One(&result)
	if err != nil {
		return 0, err
	}

	count, ok := result.(float64)
	if !ok {
		return 0, core.ErrorInternalError
	}

	return uint(count), nil
}

func (s *Store) UserExists(username, email string) error {
	// RethinkDB doesn't support unique secondary indexes, so this ugly code is needed
	// TODO: Rewrite when RethinkDB supports unique secondary indexes

	count, err := s.count("username", username)
	if err != nil {
		return err
	}

	if count != 0 {
		return core.ErrorUserAlreadyExists
	}

	count, err = s.count("email", email)
	if err != nil {
		return err
	}

	if count != 0 {
		return core.ErrorUserAlreadyExists
	}

	return nil
}

func (s *Store) Insert(user *User, password string) (id string, err error) {
	err = s.UserExists(user.Username, user.Email)
	if err != nil {
		return
	}

	// TODO: Change the cost
	user.Password, err = bcrypt.GenerateFromPassword([]byte(password), 0 /*cost*/)
	if err != nil {
		return
	}

	user.RegistrationTime = time.Now()
	user.IsVerified = false
	result, err := r.Table(table).Insert(user).RunWrite(s.session)
	if err != nil {
		return
	}

	id = result.GeneratedKeys[0]
	user.ID = id
	return
}

func (s *Store) SetPasswordWithID(id, password string) error {
	// TODO: Implement.
	return nil
}

func (s *Store) Update(user *User) error {
	return r.Table(table).Get(user.ID).Update(user).Exec(s.session)
}

func (s *Store) DeleteWithID(id string) error {
	return r.Table(table).Get(id).Delete().Exec(s.session)
}

func (s *Store) SetIsVerifiedWithID(id string) error {
	var data = map[string]interface{}{
		"isVerified": true,
	}
	return r.Table(table).Get(id).Update(data).Exec(s.session)
}
