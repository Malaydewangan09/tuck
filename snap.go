package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func cmdSnap() {
	parts := []string{}

	// git branch
	if branch := gitBranch(); branch != "" {
		parts = append(parts, "branch="+branch)
	}

	// listening ports
	if ports := listeningPorts(); len(ports) > 0 {
		parts = append(parts, "ports="+strings.Join(ports, ","))
	}

	// runtime versions
	if v := runtimeVersion("node", "--version"); v != "" {
		parts = append(parts, "node="+v)
	}
	if v := runtimeVersion("go", "version"); v != "" {
		parts = append(parts, "go="+v)
	}
	if v := runtimeVersion("python3", "--version"); v != "" && v != "Python" {
		parts = append(parts, "python="+v)
	}

	if len(parts) == 0 {
		fmt.Printf("%snothing to snap%s\n", dim, reset)
		return
	}

	text := strings.Join(parts, "  ")

	s, err := loadStore(localStorePath())
	if err != nil {
		fatal(err)
	}

	e := s.add(TypeSnap, text)
	if err := s.save(); err != nil {
		fatal(err)
	}

	fmt.Printf("%s%sSNAP%s saved  %s#%d%s\n", bold, magenta, reset, dim, e.ID, reset)
	fmt.Printf("  %s%s%s\n", dim, text, reset)
}

func gitBranch() string {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func listeningPorts() []string {
	out, err := exec.Command("lsof", "-iTCP", "-sTCP:LISTEN", "-P", "-n").Output()
	if err != nil {
		return nil
	}

	seen := map[string]bool{}
	ports := []string{}
	for _, line := range strings.Split(string(out), "\n")[1:] {
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}
		addr := fields[8]
		// extract port from *:3000 or 127.0.0.1:3000
		idx := strings.LastIndex(addr, ":")
		if idx < 0 {
			continue
		}
		port := addr[idx+1:]
		if port == "" || port == "*" {
			continue
		}
		if !seen[port] {
			seen[port] = true
			ports = append(ports, port)
		}
	}
	return ports
}

func runtimeVersion(bin, flag string) string {
	out, err := exec.Command(bin, flag).Output()
	if err != nil {
		return ""
	}
	v := strings.TrimSpace(string(out))
	// normalize: "go version go1.21.0 darwin/arm64" -> "1.21.0"
	if bin == "go" {
		fields := strings.Fields(v)
		if len(fields) >= 3 {
			v = strings.TrimPrefix(fields[2], "go")
		}
	}
	// "Python 3.11.0" -> "3.11.0"
	if strings.HasPrefix(v, "Python ") {
		v = strings.TrimPrefix(v, "Python ")
	}
	// "v20.11.0" -> "20.11.0"
	v = strings.TrimPrefix(v, "v")
	fields := strings.Fields(strings.Split(v, "\n")[0])
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}
