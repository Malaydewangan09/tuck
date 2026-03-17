package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type IndexEntry struct {
	Entry
	Project string `json:"project"`
	Dir     string `json:"dir"`
}

type GlobalIndex struct {
	Entries []IndexEntry `json:"entries"`
	path    string
}

func loadIndex() (*GlobalIndex, error) {
	path := globalIndexPath()
	idx := &GlobalIndex{path: path}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return idx, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &idx.Entries); err != nil {
		return nil, err
	}
	return idx, nil
}

func (idx *GlobalIndex) save() error {
	data, err := json.MarshalIndent(idx.Entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(idx.path, data, 0644)
}

func (idx *GlobalIndex) sync(dir string, entries []Entry) {
	// remove old entries for this dir
	fresh := []IndexEntry{}
	for _, e := range idx.Entries {
		if e.Dir != dir {
			fresh = append(fresh, e)
		}
	}
	// add current entries
	project := filepath.Base(dir)
	for _, e := range entries {
		fresh = append(fresh, IndexEntry{Entry: e, Project: project, Dir: dir})
	}
	idx.Entries = fresh
}

func (idx *GlobalIndex) search(term string) []IndexEntry {
	term = strings.ToLower(term)
	results := []IndexEntry{}
	for _, e := range idx.Entries {
		if strings.Contains(strings.ToLower(e.Text), term) {
			results = append(results, e)
		}
	}
	return results
}
