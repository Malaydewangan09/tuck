package main

import (
	"fmt"
	"strings"
	"time"
)

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	cyan   = "\033[36m"
	white  = "\033[37m"
)

func colorFor(t EntryType) string {
	switch t {
	case TypeNote:
		return cyan
	case TypeCmd:
		return green
	case TypeTodo:
		return blue
	case TypeWarn:
		return yellow
	}
	return white
}

func labelFor(t EntryType) string {
	switch t {
	case TypeNote:
		return "NOTE"
	case TypeCmd:
		return " CMD"
	case TypeTodo:
		return "TODO"
	case TypeWarn:
		return "WARN"
	}
	return "    "
}

func relativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return t.Format("Jan 2")
	}
}

func printEntries(entries []Entry) {
	if len(entries) == 0 {
		fmt.Printf("%s  nothing here yet. try: jot note \"something useful\"%s\n", dim, reset)
		return
	}

	groups := map[EntryType][]Entry{}
	order := []EntryType{TypeWarn, TypeTodo, TypeCmd, TypeNote}
	for _, e := range entries {
		groups[e.Type] = append(groups[e.Type], e)
	}

	first := true
	for _, t := range order {
		group := groups[t]
		if len(group) == 0 {
			continue
		}
		if !first {
			fmt.Println()
		}
		first = false

		color := colorFor(t)
		fmt.Printf("%s%s%s%s\n", bold, color, strings.ToUpper(string(t))+"S", reset)
		for _, e := range group {
			printEntry(e, false)
		}
	}
}

func printEntry(e Entry, showProject bool) {
	color := colorFor(e.Type)
	timeStr := fmt.Sprintf("%s%s%s", dim, relativeTime(e.CreatedAt), reset)

	prefix := ""
	if e.Type == TypeTodo {
		if e.Done {
			prefix = fmt.Sprintf("%s[x]%s ", dim, reset)
		} else {
			prefix = "[ ] "
		}
	}

	text := e.Text
	if e.Done {
		text = fmt.Sprintf("%s%s%s", dim, text, reset)
	}

	project := ""
	if showProject && e.Project != "" {
		project = fmt.Sprintf(" %s[%s]%s", dim, e.Project, reset)
	}

	fmt.Printf("  %s%s%d%s  %s%s  %s%s%s\n",
		color, bold, e.ID, reset,
		prefix, text,
		timeStr, project, reset)
}

func printSummary(entries []Entry) {
	if len(entries) == 0 {
		return
	}
	counts := map[EntryType]int{}
	for _, e := range entries {
		if e.Type == TypeTodo && e.Done {
			continue
		}
		counts[e.Type]++
	}

	parts := []string{}
	if n := counts[TypeWarn]; n > 0 {
		parts = append(parts, fmt.Sprintf("%s%d warn%s", yellow, n, reset))
	}
	if n := counts[TypeTodo]; n > 0 {
		parts = append(parts, fmt.Sprintf("%s%d todo%s", blue, n, reset))
	}
	if n := counts[TypeCmd]; n > 0 {
		parts = append(parts, fmt.Sprintf("%s%d cmd%s", green, n, reset))
	}
	if n := counts[TypeNote]; n > 0 {
		parts = append(parts, fmt.Sprintf("%s%d note%s", cyan, n, reset))
	}

	if len(parts) > 0 {
		fmt.Printf("%sjot:%s %s\n", dim, reset, strings.Join(parts, "  "))
	}
}
