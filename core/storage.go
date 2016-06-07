package core

type Storage interface {
	Init() error
	Close() error
}

type PlainTextStorage interface {
	Storage
	Get(user string) (string, error)
}
