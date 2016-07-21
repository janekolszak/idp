package verifier

import (
	// "fmt"
	r "gopkg.in/dancannon/gorethink.v2"
	"sync"
	// "time"
)

// Data received from RethinkDB in the Change feed
type VerificationChange struct {
	Old *Verification `gorethink:"old_val"`
	New *Verification `gorethink:"new_val"`
}

// Gets Verifications and sends emails.
// Should be spawned only in one instance.
type Worker struct {
	session   *r.Session
	table     string
	ctrl      chan bool
	waitGroup sync.WaitGroup
}

func NewWorker(session *r.Session) (*Worker, error) {
	w := new(Worker)
	w.session = session
	w.table = "verifyEmails"

	setupDatabase(session, w.table)

	return w, nil
}

// Start the Worker that sends the verification emails.
// VW will block waiting for new Verifications.
func (w *Worker) Start() error {
	cursor, err := r.Table(w.table).Filter(map[string]interface{}{"sentCount": 0}).Changes(r.ChangesOpts{IncludeInitial: true}).Run(w.session)
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
				// TODO: Process data
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
