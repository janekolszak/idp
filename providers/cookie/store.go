package cookie

type Store interface {
	Get(selector string) (user string, hash string, err error)
	Upsert(selector, user, hash string) (err error)
}
