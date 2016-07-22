package verifier

import (
	"bytes"
	"fmt"
	r "gopkg.in/dancannon/gorethink.v2"
	"html/template"
	"net/smtp"
	"sync"
	// "time"
)

// Data received from RethinkDB in the Change feed
type VerificationChange struct {
	Old *Verification `gorethink:"old_val"`
	New *Verification `gorethink:"new_val"`
}

type WorkerOpts struct {
	Session       *r.Session
	HostAddress   string
	Sender        string
	SmtpAuth      smtp.Auth
	EmailTemplate *template.Template
}

// Gets Verifications and sends emails.
// Should be spawned only in one instance.
type Worker struct {
	opt WorkerOpts

	table     string
	ctrl      chan bool
	waitGroup sync.WaitGroup
}

func NewWorker(opt WorkerOpts) (*Worker, error) {
	w := new(Worker)
	w.opt = opt
	w.table = "verifyEmails"

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

	w.ctrl = make(chan bool, 1)
	w.waitGroup.Add(1)
	go w.run(dataChannel)

	return nil
}

// Stops the Worker goroutine
func (w *Worker) Stop() {
	w.ctrl <- true
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

		case <-w.ctrl:
			// Stopping
			return
		default:
		}
	}
}

func (w *Worker) sendVerificationEmail(verification *Verification) error {
	var buffer bytes.Buffer
	w.opt.EmailTemplate.Execute(&buffer, verification)

	err := smtp.SendMail(w.opt.HostAddress, w.opt.SmtpAuth, w.opt.Sender, []string{verification.Email}, buffer.Bytes())
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
