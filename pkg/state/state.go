package state

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type State struct {
	Entries []Entry `json:"entries"`

	path string
}

type Entry struct {
	Name  string   `json:"name"`
	Home  string   `json:"home"`
	Paths []string `json:"paths"`
	// Detail config.Package `json:"detail"`
}

func (s *State) update() error {
	return nil
}

func (s *State) Add(entry Entry) error {
	s.Entries = append(s.Entries, entry)
	return nil
}

func Read(path string) (State, error) {
	var s State

	_, err := os.Stat(path)
	if err != nil {
		return State{path: path}, nil
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return s, err
	}

	if err := json.Unmarshal(content, &s); err != nil {
		return s, err
	}
	return s, nil
}

func (s *State) Save() error {
	f, err := os.Create(s.path)
	if err != nil {
		return err
	}
	return json.NewEncoder(f).Encode(s)
}
