package verifier

import (
	"github.com/janekolszak/idp/helpers"

	"fmt"
	r "gopkg.in/dancannon/gorethink.v2"
	"net/url"
	"sync"
	"time"
)

// Data received from RethinkDB in the Change feed
type RequestChange struct {
	Old *Request `gorethink:"old_val"`
	New *Request `gorethink:"new_val"`
}

type WorkerOpts struct {
	Session         *r.Session
	EndpointAddress string
	RequestMaxAge   time.Duration
	CleanupInterval time.Duration
	EmailerOpts     helpers.EmailerOpts
}

// Gets Requests and sends emails.
// Should be spawned only in one instance.
// Only one Worker can operate
type Worker struct {
	opt WorkerOpts

	table   string
	url     *url.URL
	emailer *helpers.Emailer

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

	w.emailer, err = helpers.NewEmailer(w.opt.EmailerOpts)
	if err != nil {
		return nil, err
	}

	return w, nil
}

// Start the Worker that sends the verification emails.
// VW will block waiting for new Requests.
func (w *Worker) Start() error {
	cursor, err := r.Table(w.table).Filter(map[string]interface{}{"sentCount": 0}).Changes(r.ChangesOpts{IncludeInitial: true}).Run(w.opt.Session)
	if err != nil {
		return err
	}
	// defer cursor.Close()

	dataChannel := make(chan RequestChange)
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

func (w *Worker) run(dataChannel <-chan RequestChange) {
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
				err := w.serve(c.New)
				if err != nil {
					fmt.Println(err.Error())
				}
			} else if c.Old != nil {
				// Shouldn't happen
			}

		case now, ok := <-w.cleanupTicker.C:
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

func (w *Worker) serve(verification *Request) error {
	parameters := url.Values{}
	parameters.Add("code", verification.ID)
	w.url.RawQuery = parameters.Encode()

	data := map[string]string{
		"Username": verification.Username,
		"Email":    verification.Email,
		"URL":      w.url.String(),
	}

	err := w.emailer.Send(verification.Email, data)
	if err != nil {
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
	_, err := r.Table(w.table).Between(r.MinVal, now.Add(-w.opt.RequestMaxAge), r.BetweenOpts{Index: "lastSentTime"}).Delete().Run(w.opt.Session)
	return err
}
