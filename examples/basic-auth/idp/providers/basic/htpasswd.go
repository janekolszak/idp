package basic

import (
	"bufio"
	"encoding/csv"
	"github.com/janekolszak/idp/core"
	"io"
	"os"
	"sync"
)

// TODO: Reload data if file changed
type Htpasswd struct {
	Hashes map[string]string
	mtx    sync.RWMutex
}

func (h *Htpasswd) Load(filename string) error {
	f, err := os.OpenFile(filename, os.O_RDONLY, os.ModeExclusive)
	if err != nil {
		return err
	}
	defer f.Close()

	r := csv.NewReader(bufio.NewReader(f))
	r.Comment = '#'
	r.Comma = ':'
	r.TrimLeadingSpace = true
	r.FieldsPerRecord = 2

	h.mtx.Lock()
	defer h.mtx.Unlock()

	h.Hashes = make(map[string]string)
	for {
		fields, err := r.Read()
		if err != nil {
			if err == io.EOF {
				return nil
			}

			return err
		}

		h.Hashes[fields[0]] = fields[1]
	}
}

func (h *Htpasswd) Get(user string) (string, error) {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	hash, ok := h.Hashes[user]
	if !ok {
		return "", core.ErrorNoSuchUser
	}

	return hash, nil
}
