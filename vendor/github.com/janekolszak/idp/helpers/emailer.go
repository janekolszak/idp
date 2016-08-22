package helpers

import (
	"github.com/nu7hatch/gouuid"
	gomail "gopkg.in/gomail.v2"
	"html/template"
	"io"
)

type EmailerOpts struct {
	Host         string
	Port         int
	User         string
	Password     string
	From         string
	Subject      string
	Domain       string
	TextTemplate *template.Template
	HtmlTemplate *template.Template
}

// Abstracts efficient email sending.
// It isn't thread safe.
type Emailer struct {
	opt EmailerOpts

	msg    *gomail.Message
	data   *map[string]string
	dialer *gomail.Dialer
}

func NewEmailer(opt EmailerOpts) (*Emailer, error) {
	e := new(Emailer)
	e.opt = opt

	// Prepare message
	e.msg = gomail.NewMessage()
	e.msg.SetHeader("From", opt.From)
	e.msg.SetHeader("Subject", opt.Subject)
	if e.opt.TextTemplate != nil {
		e.msg.AddAlternativeWriter("text/plain", func(writer io.Writer) error {
			return e.opt.TextTemplate.Execute(writer, *e.data)
		})
	}

	if e.opt.HtmlTemplate != nil {
		e.msg.AddAlternativeWriter("text/html", func(writer io.Writer) error {
			return e.opt.HtmlTemplate.Execute(writer, *e.data)
		})
	}

	// Dialer will send emails
	e.dialer = gomail.NewDialer(opt.Host, opt.Port, opt.User, opt.Password)

	return e, nil
}

func (e *Emailer) Send(to string, args map[string]string) error {
	// Set message-id to avoid spam filters
	// <[uid]@[sendingdomain.com]>
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}
	e.msg.SetHeader("Message-Id", "<"+id.String()+"@"+e.opt.Domain+">")

	e.data = &args
	e.msg.SetHeader("To", to)

	err = e.dialer.DialAndSend(e.msg)
	return err
}
