package rethinkdb

import "time"

type User struct {
	ID               string    `json:"id" gorethink:"id"`
	FirstName        string    `json:"firstName" gorethink:"firstName"`
	LastName         string    `json:"lastName" gorethink:"lastName"`
	Username         string    `json:"username" gorethink:"username"`
	Email            string    `json:"email" gorethink:"email"`
	IsVerified       bool      `json:"isVerified" gorethink:"isVerified"`
	RegistrationTime time.Time `json:"registrationTime" gorethink:"registrationTime"`
}

func (u *User) GetFirstName() string {
	return u.FirstName
}

func (u *User) GetLastName() string {
	return u.LastName
}

func (u *User) GetUsername() string {
	return u.Username
}

func (u *User) GetEmail() string {
	return u.Email
}

func (u *User) GetIsVerified() bool {
	return u.IsVerified
}

func (u *User) GetRegistrationTime() time.Time {
	return u.RegistrationTime
}
