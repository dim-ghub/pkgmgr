package main

import (
	"os/exec"
	"strings"
)

type Pacman struct{}

func (p *Pacman) Name() string { return "pacman" }

func (p *Pacman) Exists() bool {
	_, err := exec.LookPath("pacman")
	return err == nil
}

func (p *Pacman) Search(query string) ([]Package, error) {
	out, err := runOutput("pacman", "-Ss", query)
	if err != nil {
		return nil, err
	}
	return parsePacmanOutput(out), nil
}

func tokenizeLine(rest string) []string {
	var tokens []string
	var cur strings.Builder
	inBracket := false

	for _, ch := range rest {
		switch {
		case ch == '[':
			if cur.Len() > 0 {
				tokens = append(tokens, cur.String())
				cur.Reset()
			}
			inBracket = true
			cur.WriteRune(ch)
		case ch == ']':
			cur.WriteRune(ch)
			tokens = append(tokens, cur.String())
			cur.Reset()
			inBracket = false
		case ch == ' ' && !inBracket:
			if cur.Len() > 0 {
				tokens = append(tokens, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(ch)
		}
	}
	if cur.Len() > 0 {
		tokens = append(tokens, cur.String())
	}
	return tokens
}

func parsePacmanOutput(out string) []Package {
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return nil
	}

	var pkgs []Package
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if !strings.Contains(line, "/") {
			continue
		}

		slugIdx := strings.Index(line, "/")
		repo := strings.TrimSpace(line[:slugIdx])
		rest := strings.TrimSpace(line[slugIdx+1:])

		tokens := tokenizeLine(rest)
		if len(tokens) < 2 {
			continue
		}

		name := tokens[0]
		version := tokens[1]

		var size string
		installed := false

		for _, t := range tokens[2:] {
			if t == "[installed]" || t == "[Installed]" {
				installed = true
			} else if strings.HasPrefix(t, "[") && strings.HasSuffix(t, "]") {
				size = t[1 : len(t)-1]
			}
		}

		description := ""
		if i+1 < len(lines) {
			descLine := strings.TrimSpace(lines[i+1])
			if descLine != "" && !strings.Contains(descLine, "/") {
				description = descLine
				i++
			}
		}

		pkgs = append(pkgs, Package{
			Name:        name,
			Version:     version,
			Description: description,
			Source:      "pacman",
			Repo:        repo,
			Installed:   installed,
			Size:        size,
		})
	}
	return pkgs
}

func (p *Pacman) Install(pkg string, extraArgs ...string) error {
	args := []string{"-S", pkg}
	args = append(args, extraArgs...)
	return runSudoInteractive("pacman", args...)
}

func (p *Pacman) Remove(pkg string, extraArgs ...string) error {
	args := []string{"-Rs", pkg}
	args = append(args, extraArgs...)
	return runSudoInteractive("pacman", args...)
}

func (p *Pacman) Update(extraArgs ...string) error {
	args := []string{"-Syu"}
	args = append(args, extraArgs...)
	return runSudoInteractive("pacman", args...)
}
