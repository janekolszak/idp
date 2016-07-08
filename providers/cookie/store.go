package cookie

import "time"

type Store interface {
	Get(selector string) (user string, hash string, expiration time.Time, err error)
	Insert(user, hash string, expiration time.Time) (selector string, err error)
	Update(selector, user, hash string, expiration time.Time) (err error)
}
