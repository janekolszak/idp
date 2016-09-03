package cookie

import (
	"database/sql"
	"github.com/satori/go.uuid"
	"time"
)

type DBStore struct {
	db      *sql.DB
	getStmt *sql.Stmt
}

func NewDBStore(driverName, databaseSourceName string) (*DBStore, error) {
	var s = new(DBStore)

	var err error
	s.db, err = sql.Open(driverName, databaseSourceName)
	if err != nil {
		return nil, err
	}

	err = s.db.Ping()
	if err != nil {
		return nil, err
	}

	sqlStmt := `
		CREATE TABLE  IF NOT EXISTS cookieauth (selector   VARCHAR(20) NOT NULL PRIMARY KEY,
					                            validator  TEXT NOT NULL,
					                            user       TEXT NOT NULL,
					                            expiration DATETIME);`

	_, err = s.db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	// Prepare statements
	s.getStmt, err = s.db.Prepare("SELECT validator, user, expiration FROM cookieauth WHERE selector = ?")
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *DBStore) Insert(user, hash string, expiration time.Time) (selector string, err error) {

	// TODO: Database should generate the selector, but is can't be sequential
	uniqueID := uuid.NewV1()
	selector = uniqueID.String()

	tx, err := s.db.Begin()
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

	stmt, err := tx.Prepare("INSERT INTO cookieauth(selector, validator, user, expiration) values(?, ?, ?, ?)")
	if err != nil {
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(selector, hash, user, expiration)
	if err != nil {
		return
	}

	return
}

func (s *DBStore) Update(selector, user, hash string, expiration time.Time) (err error) {
	tx, err := s.db.Begin()
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

	stmt, err := tx.Prepare("UPDATE cookieauth SET validator=?, expiration =? WHERE selector=?")
	if err != nil {
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(hash, expiration, selector)

	return
}

func (s *DBStore) Get(selector string) (user, hash string, expiration time.Time, err error) {
	err = s.getStmt.QueryRow(selector).Scan(&hash, &user, &expiration)
	return
}

func (s *DBStore) DeleteSelector(selector string) (err error) {
	tx, err := s.db.Begin()
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

	stmt, err := tx.Prepare("DELETE FROM cookieauth WHERE selector=?")
	if err != nil {
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(selector)

	return
}

func (s *DBStore) DeleteUser(user string) (err error) {
	tx, err := s.db.Begin()
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

	stmt, err := tx.Prepare("DELETE FROM cookieauth WHERE user=?")
	if err != nil {
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(user)

	return
}

func (s *DBStore) Close() error {
	s.getStmt.Close()
	return s.db.Close()
}
