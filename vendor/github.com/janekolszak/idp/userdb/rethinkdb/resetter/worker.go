package resetter

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
	stopChannel   chan bool
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
	cursor, err := r.Table(w.table).Filter(map[string]interface{}{"isSent": false}).Changes(r.ChangesOpts{IncludeInitial: true}).Run(w.opt.Session)
	if err != nil {
		return err
	}
	// defer cursor.Close()

	dataChannel := make(chan RequestChange)
	cursor.Listen(dataChannel)

	w.cleanupTicker = time.NewTicker(w.opt.CleanupInterval)

	w.stopChannel = make(chan bool, 1)
	w.waitGroup.Add(1)
	go w.run(dataChannel)

	return nil
}

// Stops the Worker goroutine
func (w *Worker) Stop() {
	w.stopChannel <- true
	w.cleanupTicker.Stop()
	w.waitGroup.Wait()
}

func (w *Worker) run(dataChannel <-chan RequestChange) {
	defer close(w.stopChannel)
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
		case <-w.stopChannel:
			// Stopping
			return
		default:
		}
	}
}

func (w *Worker) serve(request *Request) error {
	parameters := url.Values{}
	parameters.Add("code", request.ID)
	w.url.RawQuery = parameters.Encode()

	data := map[string]string{
		"Username": request.Username,
		"Email":    request.Email,
		"URL":      w.url.String(),
	}

	err := w.emailer.Send(request.Email, data)
	if err != nil {
		return err
	}

	// Mark as sent
	_, err = r.Table(w.table).Get(request.ID).Update(map[string]interface{}{"isSent": true}).RunWrite(w.opt.Session)

	return err
}

// Deletes expired entries
func (w *Worker) deleteExpired(now time.Time) error {
	_, err := r.Table(w.table).Between(r.MinVal, now.Add(-w.opt.RequestMaxAge), r.BetweenOpts{Index: "toc"}).Delete().Run(w.opt.Session)
	return err
}
