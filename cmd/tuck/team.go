package main

import (
	"crypto/md5"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// teamFile marks that this project has team mode enabled
func teamMarkerPath() string {
	dir, _ := os.Getwd()
	return filepath.Join(dir, ".tuck-team")
}

func isTeamEnabled() bool {
	_, err := os.Stat(teamMarkerPath())
	return err == nil
}

func cmdTeam(args []string) {
	if len(args) == 0 {
		cmdTeamStatus()
		return
	}

	switch args[0] {
	case "on", "enable":
		cmdTeamOn()
	case "off", "disable":
		cmdTeamOff()
	case "sync":
		cmdTeamSync()
	default:
		fmt.Fprintf(os.Stderr, "usage: tuck team [on|off|sync]\n")
		os.Exit(1)
	}
}

func cmdTeamStatus() {
	if isTeamEnabled() {
		fmt.Printf("%steam mode: %s%son%s\n", dim, reset, green, reset)
		fmt.Printf("%s.tuck is tracked by git. run: tuck team sync after pulling.%s\n", dim, reset)
	} else {
		fmt.Printf("%steam mode: %s%soff%s\n", dim, reset, dim, reset)
		fmt.Printf("%srun: tuck team on to share notes with teammates via git.%s\n", dim, reset)
	}
}

func cmdTeamOn() {
	// create marker
	if err := os.WriteFile(teamMarkerPath(), []byte(""), 0644); err != nil {
		fatal(err)
	}

	// stage .tuck for git if it exists
	storePath := localStorePath()
	if _, err := os.Stat(storePath); err == nil {
		exec.Command("git", "add", storePath).Run()
	}

	// remove .tuck from .gitignore if present
	removeFromGitignore(".tuck")

	fmt.Printf("%s%steam mode on%s\n", bold, green, reset)
	fmt.Printf("%scommit .tuck to share notes with teammates.%s\n", dim, reset)
	fmt.Printf("%srun tuck team sync after pulling to merge their notes.%s\n", dim, reset)
}

func cmdTeamOff() {
	if err := os.Remove(teamMarkerPath()); err != nil && !os.IsNotExist(err) {
		fatal(err)
	}
	fmt.Printf("%steam mode off%s\n", dim, reset)
}

func cmdTeamSync() {
	// load current store
	s, err := loadStore(localStorePath())
	if err != nil {
		fatal(err)
	}

	// get the git version of .tuck (from HEAD)
	out, err := exec.Command("git", "show", "HEAD:.tuck").Output()
	if err != nil {
		fmt.Printf("%snothing to sync (no committed .tuck found)%s\n", dim, reset)
		return
	}

	// parse git version into a temporary store
	tmpPath := localStorePath() + ".tmp"
	if err := os.WriteFile(tmpPath, out, 0644); err != nil {
		fatal(err)
	}
	defer os.Remove(tmpPath)

	remote, err := loadStore(tmpPath)
	if err != nil {
		fatal(err)
	}

	// merge: add entries from remote that don't exist locally (by content hash)
	added := 0
	for _, re := range remote.Entries {
		if !hasEntry(s.Entries, re) {
			s.add(re.Type, re.Text)
			added++
		}
	}

	if added == 0 {
		fmt.Printf("%salready up to date%s\n", dim, reset)
		return
	}

	if err := s.save(); err != nil {
		fatal(err)
	}

	dir, _ := os.Getwd()
	idx, _ := loadIndex()
	if idx != nil {
		idx.sync(dir, s.Entries)
		idx.save()
	}

	fmt.Printf("%s%s+%d%s entries synced from teammates\n", bold, green, added, reset)
}

// hasEntry checks by content hash to avoid duplicates
func hasEntry(entries []Entry, e Entry) bool {
	h := entryHash(e)
	for _, existing := range entries {
		if entryHash(existing) == h {
			return true
		}
	}
	return false
}

func entryHash(e Entry) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(string(e.Type)+e.Text)))
}

func removeFromGitignore(pattern string) {
	dir, _ := os.Getwd()
	path := filepath.Join(dir, ".gitignore")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	filtered := []string{}
	for _, l := range lines {
		if strings.TrimSpace(l) != pattern {
			filtered = append(filtered, l)
		}
	}
	os.WriteFile(path, []byte(strings.Join(filtered, "\n")), 0644)
}
