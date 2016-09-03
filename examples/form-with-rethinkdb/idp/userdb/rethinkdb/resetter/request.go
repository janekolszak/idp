package resetter

import (
	r "gopkg.in/dancannon/gorethink.v2"
	"time"
)

type Request struct {
	// ID is a random UUIDv4. We will use it as the "selector code"
	ID string `json:"id,omitempty" gorethink:"id,omitempty"`

	// User whose email we validate
	UserID string `json:"userID" gorethink:"userID"`

	// Username will be passed to an email template
	Username string `json:"username" gorethink:"username"`

	// Which email to verify
	Email string `json:"email" gorethink:"email"`

	// This field holds how many times the verification email was resent
	IsSent bool `json:"isSent" gorethink:"isSent"`

	// Time of creation
	TOC time.Time `json:"toc" gorethink:"toc"`
}

func setupDatabase(session *r.Session, table string) {
	// Discard error (database exists)
	db := session.Database()
	r.DBCreate(db).RunWrite(session)
	r.DB(db).TableCreate(table).RunWrite(session)

	// Index for fetching users by credentials used in the login
	r.Table(table).IndexCreate("toc").Exec(session)

	r.Table(table).IndexWait().RunWrite(session)
}
