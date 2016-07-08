package cookie

import (
	"database/sql"
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

func (s *DBStore) Upsert(selector, user, hash string, expiration time.Time) (err error) {
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

func (s *DBStore) Get(selector string) (user, hash string, expiration time.Time, err error) {
	err = s.getStmt.QueryRow(selector).Scan(&hash, &user, &expiration)
	return
}

func (s *DBStore) Close() error {
	s.getStmt.Close()
	return s.db.Close()
}
