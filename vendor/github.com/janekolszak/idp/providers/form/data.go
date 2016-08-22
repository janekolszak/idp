package form

import (
	"github.com/janekolszak/idp/core"

	"github.com/asaskevich/govalidator"
	"github.com/gorilla/schema"
	"net/http"
)

var (
	decoder = schema.NewDecoder()
)

type RegisterPOST struct {
	Password  string `schema:"password"  valid:"required,password"`
	Username  string `schema:"username"  valid:"required,ascii,length(1|70)"`
	Email     string `schema:"email"     valid:"required,email,length(1|100)"`
	FirstName string `schema:"firstName" valid:"required,ascii,length(1|35)"`
	LastName  string `schema:"lastName"  valid:"required,ascii,length(1|35)"`
}

func NewRegisterPOST(r *http.Request) (*RegisterPOST, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}

	data := new(RegisterPOST)
	err = decoder.Decode(data, r.PostForm)
	if err != nil {
		return nil, err
	}

	valid, err := govalidator.ValidateStruct(data)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, core.ErrorBadRequest
	}

	return data, nil
}

type VerifyGET struct {
	Code string `schema:"code"  valid:"required,uuidv4"`
}

func NewVerifyGET(r *http.Request) (*VerifyGET, error) {
	data := new(VerifyGET)
	err := decoder.Decode(data, r.URL.Query())
	if err != nil {
		return nil, err
	}

	valid, err := govalidator.ValidateStruct(data)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, core.ErrorBadRequest
	}

	return data, nil
}
