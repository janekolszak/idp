package helpers

import (
	"../core"
	"bufio"
	"encoding/csv"
	"io"
	"os"
)

// TODO: Reload data if file changed
type Htpasswd struct {
	Hashes map[string]string
}

func (h *Htpasswd) load(filename string) error {
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

func NewHtpasswd(filename string) (*Htpasswd, error) {
	h := new(Htpasswd)

	err := h.load(filename)
	if err != nil {
		return nil, err
	}
	return h, nil
}

func (h *Htpasswd) Get(user string) (string, error) {
	hash, ok := h.Hashes[user]
	if !ok {
		return "", core.ErrorNoSuchUser
	}

	return hash, nil
}
