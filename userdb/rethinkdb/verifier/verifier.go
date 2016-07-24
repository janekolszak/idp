package verifier

import (
	"github.com/janekolszak/idp/core"

	// "fmt"
	r "gopkg.in/dancannon/gorethink.v2"
)

// Verifier sends user a request email with a random code.
// (UUID v4 according to https://www.rethinkdb.com/api/javascript/uuid/ )
// When user returns with this particular code he's considered to be verified.https://www.rethinkdb.com/api/javascript/uuid/
type Verifier struct {
	session *r.Session
	table   string
}

func NewVerifier(session *r.Session) (*Verifier, error) {
	v := new(Verifier)
	v.session = session
	v.table = "verifyEmails"

	setupDatabase(session, v.table)

	return v, nil
}

// Pushes the request to the database.
// Later a Verify Worker will pop the request and send the email.
func (v *Verifier) Push(userID, username, email string) (code string, err error) {
	request := Request{
		UserID:    userID,
		Username:  username,
		Email:     email,
		SentCount: 0,
	}

	// TODO: Check result
	resp, err := r.Table(v.table).Insert(request).RunWrite(v.session)
	return resp.GeneratedKeys[0], err
}

// Called from the http handler when the request code is received.
// Gets the user id assigned to the code and removes the underlying Request from the database.
func (v *Verifier) Verify(code string) (userID string, err error) {
	resp, err := r.Table(v.table).Get(code).Delete(r.DeleteOpts{ReturnChanges: true}).RunWrite(v.session)
	if err != nil {
		return
	}

	// fmt.Println(resp.Changes[0])
	request := resp.Changes[0].OldValue.(map[string]interface{})
	userID = request["userID"].(string)
	return
}

func (v *Verifier) Count() (uint, error) {
	cursor, err := r.Table(v.table).Count().Run(v.session)
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
