package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type EntryType string

const (
	TypeNote EntryType = "note"
	TypeCmd  EntryType = "cmd"
	TypeTodo EntryType = "todo"
	TypeWarn EntryType = "warn"
)

type Entry struct {
	ID        int       `json:"id"`
	Type      EntryType `json:"type"`
	Text      string    `json:"text"`
	Done      bool      `json:"done,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	Project   string    `json:"project,omitempty"` // used in global index
	Dir       string    `json:"dir,omitempty"`     // used in global index
}

type Store struct {
	Entries []Entry `json:"entries"`
	path    string
}

func localStorePath() string {
	dir, _ := os.Getwd()
	return filepath.Join(dir, ".tuck")
}

func globalIndexPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".tuck-index")
}

func loadStore(path string) (*Store, error) {
	s := &Store{path: path}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &s.Entries); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) save() error {
	data, err := json.MarshalIndent(s.Entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

func (s *Store) add(t EntryType, text string) Entry {
	id := 1
	for _, e := range s.Entries {
		if e.ID >= id {
			id = e.ID + 1
		}
	}
	e := Entry{
		ID:        id,
		Type:      t,
		Text:      text,
		CreatedAt: time.Now(),
	}
	s.Entries = append(s.Entries, e)
	return e
}

func (s *Store) remove(id int) bool {
	for i, e := range s.Entries {
		if e.ID == id {
			s.Entries = append(s.Entries[:i], s.Entries[i+1:]...)
			return true
		}
	}
	return false
}

func (s *Store) get(id int) *Entry {
	for i, e := range s.Entries {
		if e.ID == id {
			return &s.Entries[i]
		}
	}
	return nil
}

func (s *Store) toggleDone(id int) bool {
	for i, e := range s.Entries {
		if e.ID == id {
			s.Entries[i].Done = !e.Done
			return true
		}
	}
	return false
}
