package cookie

import "time"

type Store interface {
	Get(selector string) (user string, hash string, expiration time.Time, err error)
	Upsert(selector, user, hash string, expiration time.Time) (err error)
}
