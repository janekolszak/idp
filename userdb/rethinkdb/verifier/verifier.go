package verifier

import (
	"github.com/janekolszak/idp/core"

	// "fmt"
	r "gopkg.in/dancannon/gorethink.v2"
	"time"
)

type Verification struct {
	// ID is a random UUIDv4. We will use it as the "selector code"
	ID string `json:"id,omitempty" gorethink:"id,omitempty"`

	// User whose email we validate
	UserID string `json:"userID" gorethink:"userID"`

	// Username will be passed to an email template
	Username string `json:"username" gorethink:"username"`

	// Which email to verify
	Email string `json:"email" gorethink:"email"`

	// This field holds how many times the verification email was resent
	SentCount int `json:"sentCount" gorethink:"sentCount"`

	// Time of sending the last email
	LastSentTime time.Time `json:"lastSentTime,omitempty" gorethink:"lastSentTime,omitempty"`
}

// Verifier sends user a verification email with a random code.
// (UUID v4 according to https://www.rethinkdb.com/api/javascript/uuid/ )
// When user returns with this particular code he's considered to be verified.https://www.rethinkdb.com/api/javascript/uuid/
type Verifier struct {
	session *r.Session
	table   string
}

func setupDatabase(session *r.Session, table string) {
	// Discard error (database exists)
	db := session.Database()
	r.DBCreate(db).RunWrite(session)
	r.DB(db).TableCreate(table).RunWrite(session)

	// Index for fetching users by credentials used in the login
	r.Table(table).IndexCreate("lastSentTime").Exec(session)

	r.Table(table).IndexWait().RunWrite(session)
}

func NewVerifier(session *r.Session) (*Verifier, error) {
	v := new(Verifier)
	v.session = session
	v.table = "verifyEmails"

	setupDatabase(session, v.table)

	return v, nil
}

// Pushes the verification to the database.
// Later a Verify Worker will pop the verification and send the email.
func (v *Verifier) PushVerification(userID, username, email string) (code string, err error) {
	verification := Verification{
		UserID:    userID,
		Username:  username,
		SentCount: 0,
	}

	// TODO: Check result
	resp, err := r.Table(v.table).Insert(verification).RunWrite(v.session)
	return resp.GeneratedKeys[0], err
}

// Called from the http handler when the verification code is received.
// Gets the user id assigned to the code and removes the underlying Verification from the database.
func (v *Verifier) Verify(code string) (userID string, err error) {
	resp, err := r.Table(v.table).Get(code).Delete(r.DeleteOpts{ReturnChanges: true}).RunWrite(v.session)
	if err != nil {
		return
	}

	// fmt.Println(resp.Changes[0])
	verification := resp.Changes[0].OldValue.(map[string]interface{})
	userID = verification["userID"].(string)
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
