package resetter

import (
	"github.com/janekolszak/idp/core"

	// "fmt"
	r "gopkg.in/dancannon/gorethink.v2"
)

// Password Reseter manages password reset requests.
type Resetter struct {
	session *r.Session
	table   string
}

func NewResetter(session *r.Session) (*Resetter, error) {
	resetter := new(Resetter)
	resetter.session = session
	resetter.table = "passwordReset"

	setupDatabase(session, resetter.table)

	return resetter, nil
}

// Pushes the request to the database.
// Later a Worker will pop the verification and send the email.
func (resetter *Resetter) Push(userID, username, email string) (code string, err error) {
	req := Request{
		UserID:   userID,
		Username: username,
		Email:    email,
		IsSent:   false,
	}

	// TODO: Check result
	resp, err := r.Table(resetter.table).Insert(req).RunWrite(resetter.session)
	return resp.GeneratedKeys[0], err
}

// Called from the http handler when the reset password request is received
// Gets the user id assigned to the code but leaves the request intact.
func (resetter *Resetter) GetWithCode(code string) (userID string, err error) {
	cursor, err := r.Table(resetter.table).Get(code).Pluck("userID").Run(resetter.session)
	if err != nil {
		return
	}
	defer cursor.Close()

	if cursor.IsNil() {
		err = core.ErrorNoSuchEntry
		return
	}

	var data map[string]string
	err = cursor.One(&data)
	if err != nil {
		return
	}

	userID = data["userID"]
	return
}

// Called from the http handler when the verification code is received.
// Gets the user id assigned to the code and removes the underlying Request from the database.
func (resetter *Resetter) Reset(code string) (userID string, err error) {
	resp, err := r.Table(resetter.table).Get(code).Delete(r.DeleteOpts{ReturnChanges: true}).RunWrite(resetter.session)
	if err != nil {
		return
	}

	verification := resp.Changes[0].OldValue.(map[string]interface{})
	userID = verification["userID"].(string)
	return
}

func (resetter *Resetter) Count() (uint, error) {
	cursor, err := r.Table(resetter.table).Count().Run(resetter.session)
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
