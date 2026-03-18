package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const version = "0.2.0"
const appName = "tuck"

func printHelp() {
	fmt.Printf("\n  %s%stuck%s  %snothing gets forgotten%s  %sv%s%s\n\n", bold, white, reset, dim, reset, dim, version, reset)

	fmt.Printf("  %sSAVE%s\n", bold, reset)
	fmt.Printf("  tuck %snote%s <text>    save a note\n", cyan, reset)
	fmt.Printf("  tuck %scmd%s  <text>    save a runnable command\n", green, reset)
	fmt.Printf("  tuck %stodo%s <text>    save a todo item\n", blue, reset)
	fmt.Printf("  tuck %swarn%s <text>    save a warning\n", yellow, reset)
	fmt.Printf("  tuck %ssnap%s           snapshot branch, ports, runtimes\n", magenta, reset)

	fmt.Printf("\n  %sVIEW & ACT%s\n", bold, reset)
	fmt.Printf("  tuck %sls%s             list entries for current project\n", dim, reset)
	fmt.Printf("  tuck %srun%s <id>       execute a saved command\n", dim, reset)
	fmt.Printf("  tuck %sdone%s <id>      toggle todo done / undone\n", dim, reset)
	fmt.Printf("  tuck %srm%s <id>        remove an entry\n", dim, reset)
	fmt.Printf("  tuck %sgrep%s <term>    search across all projects\n", dim, reset)

	fmt.Printf("\n  %sTEAM%s\n", bold, reset)
	fmt.Printf("  tuck %steam on%s        share notes via git\n", dim, reset)
	fmt.Printf("  tuck %steam sync%s      merge teammates notes after pull\n", dim, reset)

	fmt.Printf("\n")
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		printHelp()
		return
	}

	cmd := args[0]
	rest := args[1:]

	switch cmd {
	case "note", "cmd", "todo", "warn":
		cmdAdd(EntryType(cmd), rest)
	case "ls", "list":
		cmdList()
	case "run":
		cmdRun(rest)
	case "done", "check":
		cmdDone(rest)
	case "rm", "remove", "delete":
		cmdRemove(rest)
	case "clear":
		cmdClear()
	case "grep", "search", "find":
		cmdGrep(rest)
	case "summary":
		cmdSummary()
	case "snap":
		cmdSnap()
	case "team":
		cmdTeam(rest)
	case "version", "--version", "-v":
		fmt.Printf("tuck %s\n", version)
	case "help", "--help", "-h":
		printHelp()
	default:
		// treat as shorthand: tuck "text" → tuck note "text"
		if len(cmd) > 0 && cmd[0] != '-' {
			cmdAdd(TypeNote, args)
		} else {
			printHelp()
		}
	}
}

func cmdAdd(t EntryType, args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "error: provide some text\n")
		os.Exit(1)
	}
	text := strings.Join(args, " ")

	s, err := loadStore(localStorePath())
	if err != nil {
		fatal(err)
	}

	e := s.add(t, text)
	if err := s.save(); err != nil {
		fatal(err)
	}

	// sync global index
	dir, _ := os.Getwd()
	idx, _ := loadIndex()
	if idx != nil {
		idx.sync(dir, s.Entries)
		idx.save()
	}

	color := colorFor(t)
	fmt.Printf("\n  %s%s %s%s  %s#%d  %s%s\n\n", bold, color, strings.ToUpper(string(t)), reset, dim, e.ID, relativeTime(e.CreatedAt), reset)
}

func cmdList() {
	s, err := loadStore(localStorePath())
	if err != nil {
		fatal(err)
	}

	dir, _ := os.Getwd()
	project := filepath.Base(dir)
	fmt.Printf("\n  %s%s%s  %s%s%s\n", bold, project, reset, dim, dir, reset)

	printEntries(s.Entries)
}

func cmdRun(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "error: provide an entry id\n")
		os.Exit(1)
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid id\n")
		os.Exit(1)
	}

	s, err := loadStore(localStorePath())
	if err != nil {
		fatal(err)
	}

	e := s.get(id)
	if e == nil {
		fmt.Fprintf(os.Stderr, "error: entry #%d not found\n", id)
		os.Exit(1)
	}
	if e.Type != TypeCmd {
		fmt.Fprintf(os.Stderr, "error: entry #%d is not a command\n", id)
		os.Exit(1)
	}

	fmt.Printf("%s$ %s%s\n", dim, e.Text, reset)
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	c := exec.Command(shell, "-c", e.Text)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		os.Exit(1)
	}
}

func cmdDone(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "error: provide an entry id\n")
		os.Exit(1)
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid id\n")
		os.Exit(1)
	}

	s, err := loadStore(localStorePath())
	if err != nil {
		fatal(err)
	}

	if !s.toggleDone(id) {
		fmt.Fprintf(os.Stderr, "error: entry #%d not found\n", id)
		os.Exit(1)
	}

	if err := s.save(); err != nil {
		fatal(err)
	}

	e := s.get(id)
	if e.Done {
		fmt.Printf("%s%s[x] done%s\n", green, bold, reset)
	} else {
		fmt.Printf("%s%s[ ] undone%s\n", dim, bold, reset)
	}
}

func cmdRemove(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "error: provide an entry id\n")
		os.Exit(1)
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid id\n")
		os.Exit(1)
	}

	s, err := loadStore(localStorePath())
	if err != nil {
		fatal(err)
	}

	if !s.remove(id) {
		fmt.Fprintf(os.Stderr, "error: entry #%d not found\n", id)
		os.Exit(1)
	}

	if err := s.save(); err != nil {
		fatal(err)
	}

	// sync index
	dir, _ := os.Getwd()
	idx, _ := loadIndex()
	if idx != nil {
		idx.sync(dir, s.Entries)
		idx.save()
	}

	fmt.Printf("%sremoved #%d%s\n", dim, id, reset)
}

func cmdClear() {
	path := localStorePath()
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		fatal(err)
	}

	dir, _ := os.Getwd()
	idx, _ := loadIndex()
	if idx != nil {
		idx.sync(dir, nil)
		idx.save()
	}

	fmt.Printf("%scleared%s\n", dim, reset)
}

func cmdGrep(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "error: provide a search term\n")
		os.Exit(1)
	}
	term := strings.Join(args, " ")

	idx, err := loadIndex()
	if err != nil {
		fatal(err)
	}

	results := idx.search(term)
	if len(results) == 0 {
		fmt.Printf("%sno results for \"%s\"%s\n", dim, term, reset)
		return
	}

	// group by project
	seen := map[string]bool{}
	projects := []string{}
	byProject := map[string][]IndexEntry{}
	for _, e := range results {
		if !seen[e.Dir] {
			seen[e.Dir] = true
			projects = append(projects, e.Dir)
		}
		byProject[e.Dir] = append(byProject[e.Dir], e)
	}

	for i, dir := range projects {
		if i > 0 {
			fmt.Println()
		}
		project := filepath.Base(dir)
		fmt.Printf("%s%s%s  %s%s%s\n", bold, project, reset, dim, dir, reset)
		for _, e := range byProject[dir] {
			printEntry(e.Entry, false)
		}
	}
}

func cmdSummary() {
	s, err := loadStore(localStorePath())
	if err != nil {
		return
	}
	printSummary(s.Entries)
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
