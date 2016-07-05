package cookie

import (
	"github.com/janekolszak/idp/core"
	"github.com/janekolszak/idp/helpers"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"os"
)

const (
	rememberMeCookieName = "remember"
)

type CookieAuth struct {
	DB *sql.DB
}

func NewCookieAuth(filename string) (*CookieAuth, error) {
	var c = new(CookieAuth)

	isDatabaseReady := false
	if _, err := os.Stat(filename); err == nil {
		isDatabaseReady = true
	}

	// Setup database
	var err error
	c.DB, err = sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	if !isDatabaseReady {
		// There was no file before

		sqlStmt := `
		CREATE TABLE cookieauth (selector  VARCHAR(20) NOT NULL PRIMARY KEY,
	                             validator TEXT NOT NULL,
	                             user      TEXT NOT NULL);`

		_, err = c.DB.Exec(sqlStmt)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *CookieAuth) Close() error {
	return c.DB.Close()
}

func (c *CookieAuth) Check(r *http.Request) (user string, err error) {
	l, err := helpers.GetLoginCookie(r, rememberMeCookieName)
	if err != nil {
		return
	}

	// TODO: Validate selector, shouldn't be too long etc.

	// Get the credentials pointed by selector from the cookie
	stmt, err := c.DB.Prepare("SELECT validator, user FROM cookieauth WHERE selector = ?")
	if err != nil {
		return
	}
	defer stmt.Close()

	var hash string
	// TODO: Does QueryRow escape the string to avoid sql injection?
	err = stmt.QueryRow(l.Selector).Scan(&hash, &user)
	if err != nil {
		// Probably no such selector
		return
	}

	if !l.Check(hash) {
		err = core.ErrorBadRequest
	}

	return
}

func (c *CookieAuth) WriteError(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (c *CookieAuth) Write(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (c *CookieAuth) saveToDB(selector, hash, user string) (err error) {
	tx, err := c.DB.Begin()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	stmt, err := tx.Prepare("INSERT INTO cookieauth(selector, validator, user) values(?, ?, ?)")
	if err != nil {
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(selector, hash, user)
	if err != nil {
		return
	}

	return
}

// TODO: Selector should be created by the database, here it's automatically generated
func (c *CookieAuth) Add(w http.ResponseWriter, r *http.Request, user string) (err error) {
	l, err := helpers.NewLoginCookie("", rememberMeCookieName)
	if err != nil {
		return
	}

	hash, err := l.GenerateValidator()
	if err != nil {
		return
	}

	// First save to the database
	err = c.saveToDB(l.Selector, hash, user)
	if err != nil {
		return
	}

	// Then save to the cookie
	err = l.Save(w, r)
	return
}
