package providers

import (
	"github.com/janekolszak/idp/core"
	"github.com/janekolszak/idp/helpers"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
)

const (
	rememberMeCookieName = "remember"
)

type CookieAuth struct {
	DB *sql.DB
}

func NewCookieAuth(filename string) (*CookieAuth, error) {
	var c = new(CookieAuth)

	// Setup database
	var err error
	c.DB, err = sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	sqlStmt := `
	CREATE TABLE cookieauth (selector  VARCHAR(20) NOT NULL PRIMARY KEY, 
	                         validator TEXT NOT NULL,
	                         user      TEXT NOT NULL);`

	_, err = c.DB.Exec(sqlStmt)
	if err != nil {
		return nil, err
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

func (c *CookieAuth) Respond(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (c *CookieAuth) Add(user string) error {
	tx, err := c.DB.Begin()
	if err != nil {
		return err
	}

	l, err := helpers.NewLoginCookie("", "remember")
	if err != nil {
		return err
	}

	hash, err := l.GenerateValidator()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO cookieauth(selector, validator, user) values(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(l.Selector, hash, user)
	if err != nil {
		return err
	}

	return tx.Commit()
}
