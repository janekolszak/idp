package memory

import (
	"bufio"
	"encoding/csv"
	"io"
	"os"
	"sync"

	"github.com/janekolszak/idp/core"
	"golang.org/x/crypto/bcrypt"
)

type Store struct {
	hashes map[string]string
	mtx    sync.RWMutex
}

func NewMemStore() (*Store, error) {
	s := Store{}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.hashes = make(map[string]string)

	return &s, nil
}

func (s *Store) LoadHtpasswd(filename string) error {

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

	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.hashes = make(map[string]string)
	for {
		fields, err := r.Read()
		if err != nil {
			if err == io.EOF {
				return nil
			}

			return err
		}

		s.hashes[fields[0]] = fields[1]
	}
}

func (s *Store) Check(username, password string) error {

	s.mtx.RLock()
	defer s.mtx.RUnlock()

	hash, exists := s.hashes[username]
	// possibly compare against zero hash to prevent timing attack
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if !exists {
		return core.ErrorNoSuchUser
	}

	if err != nil {
		return core.ErrorAuthenticationFailure
	}

	return nil
}

// TODO: add complexity requirements
func (s *Store) Add(username, password string) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	_, exists := s.hashes[username]
	if exists {
		return core.ErrorUserAlreadyExists
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	s.hashes[username] = string(hash)
	return nil
}
