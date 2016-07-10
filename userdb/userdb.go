package userdb

type Store interface {
	Check(username, password string) error
	Add(username, password string) error
}

type UserInfo interface {
	GetFirstName() string
	GetLastName() string
	GetUsername() string
	GetEmail() string
	GetIsVerified() bool
	GetRegistrationTime() time.Time
}

type UserStore interface {
	Get(username string) (UserInfo, error)
	Check(username, password string) error
	Insert(userinfo UserInfo) error
	Update(userinfo UserInfo) error
	Delete(username string) error
}
