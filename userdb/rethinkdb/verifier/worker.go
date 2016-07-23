package verifier

import (
	"fmt"
	"github.com/nu7hatch/gouuid"
	r "gopkg.in/dancannon/gorethink.v2"
	gomail "gopkg.in/gomail.v2"
	"html/template"
	"io"
	"net/url"
	"sync"
	"time"
)

// Data received from RethinkDB in the Change feed
type VerificationChange struct {
	Old *Verification `gorethink:"old_val"`
	New *Verification `gorethink:"new_val"`
}

type WorkerOpts struct {
	Session            *r.Session
	Host               string
	Port               int
	User               string
	Password           string
	From               string
	Subject            string
	EndpointAddress    string
	TextTemplate       *template.Template
	HtmlTemplate       *template.Template
	VerificationMaxAge time.Duration
	CleanupInterval    time.Duration
}

// Gets Verifications and sends emails.
// Should be spawned only in one instance.
// Only one Worker can operate
type Worker struct {
	opt WorkerOpts

	table  string
	url    *url.URL
	msg    *gomail.Message
	data   map[string]string
	dialer *gomail.Dialer

	// Process control
	ctrl          chan bool
	cleanupTicker *time.Ticker
	waitGroup     sync.WaitGroup
}

func NewWorker(opt WorkerOpts) (*Worker, error) {
	w := new(Worker)
	w.opt = opt
	w.table = "verifyEmails"

	var err error
	w.url, err = url.Parse(opt.EndpointAddress)
	if err != nil {
		return nil, err
	}

	// Prepare message
	w.msg = gomail.NewMessage()
	w.msg.SetHeader("From", opt.From)
	w.msg.SetHeader("Subject", opt.Subject)
	if w.opt.TextTemplate != nil {
		w.msg.AddAlternativeWriter("text/plain", func(writer io.Writer) error {
			return w.opt.TextTemplate.Execute(writer, w.data)
		})
	}

	if w.opt.HtmlTemplate != nil {
		w.msg.AddAlternativeWriter("text/html", func(writer io.Writer) error {
			return w.opt.HtmlTemplate.Execute(writer, w.data)
		})
	}

	// Dialer will send emails
	w.dialer = gomail.NewDialer(opt.Host, opt.Port, opt.User, opt.Password)
	setupDatabase(w.opt.Session, w.table)

	return w, nil
}

// Start the Worker that sends the verification emails.
// VW will block waiting for new Verifications.
func (w *Worker) Start() error {
	cursor, err := r.Table(w.table).Filter(map[string]interface{}{"sentCount": 0}).Changes(r.ChangesOpts{IncludeInitial: true}).Run(w.opt.Session)
	if err != nil {
		return err
	}
	// defer cursor.Close()

	dataChannel := make(chan VerificationChange)
	cursor.Listen(dataChannel)

	w.cleanupTicker = time.NewTicker(w.opt.CleanupInterval)

	w.ctrl = make(chan bool, 1)
	w.waitGroup.Add(1)
	go w.run(dataChannel)

	return nil
}

// Stops the Worker goroutine
func (w *Worker) Stop() {
	w.ctrl <- true
	w.cleanupTicker.Stop()
	w.waitGroup.Wait()
}

func (w *Worker) run(dataChannel <-chan VerificationChange) {
	defer close(w.ctrl)
	defer w.waitGroup.Done()

	for {
		select {
		case c, ok := <-dataChannel:
			if !ok {
				// For example database closing
				return
			}
			if c.New != nil {
				// TODO: Check error. Should probably close on some errors.
				err := w.sendVerificationEmail(c.New)
				if err != nil {
					fmt.Println(err.Error())
				}
			} else if c.Old != nil {
				// Shouldn't happen
			}

		case now, ok := <-w.cleanupTicker.C:
			fmt.Println(now, ok)
			if !ok {
				// I have no idea why..
				return
			}
			err := w.deleteExpired(now)
			if err != nil {
				fmt.Println(err.Error())
			}
		case <-w.ctrl:
			// Stopping
			return
		default:
		}
	}
}

func (w *Worker) sendVerificationEmail(verification *Verification) error {
	parameters := url.Values{}
	parameters.Add("code", verification.ID)
	w.url.RawQuery = parameters.Encode()

	w.msg.SetHeader("To", verification.Email)

	id, err := uuid.NewV4()
	if err != nil {
		return err
	}
	// <[uid]@[sendingdomain.com]>
	w.msg.SetHeader("Message-Id", "<"+id.String()+"@"+w.url.Host+">")

	w.data = map[string]string{
		"Username": verification.Username,
		"Email":    verification.Email,
		"URL":      w.url.String(),
	}

	err = w.dialer.DialAndSend(w.msg)
	if err != nil {
		// Failed to send the email
		return err
	}

	// Increment sent emails counter
	_, err = r.Table(w.table).Get(verification.ID).Update(map[string]interface{}{
		"sentCount":    r.Row.Field("sentCount").Add(1).Default(0),
		"lastSentTime": r.Now()}).RunWrite(w.opt.Session)

	return err
}

// Deletes expired entries
func (w *Worker) deleteExpired(now time.Time) error {
	_, err := r.Table(w.table).Between(r.MinVal, now.Add(-w.opt.VerificationMaxAge), r.BetweenOpts{Index: "lastSentTime"}).Delete().Run(w.opt.Session)
	return err
}
