package main

import (
	"os/exec"
	"strings"
)

type AUR struct{}

func (a *AUR) Name() string { return "aur" }

func (a *AUR) Exists() bool {
	_, err := exec.LookPath("paru")
	return err == nil
}

func (a *AUR) Search(query string) ([]Package, error) {
	out, err := runOutput("paru", "-Ssa", query)
	if err != nil {
		return nil, err
	}
	return parseAUROutput(out), nil
}

func parseAUROutput(out string) []Package {
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

		installed := false

		for _, t := range tokens[2:] {
			if t == "[Installed]" {
				installed = true
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
			Source:      "aur",
			Repo:        repo,
			Installed:   installed,
		})
	}
	return pkgs
}

func (a *AUR) Install(pkg string, extraArgs ...string) error {
	args := []string{"-Sa", pkg}
	args = append(args, extraArgs...)
	return runInteractive("paru", args...)
}

func (a *AUR) Remove(pkg string, extraArgs ...string) error {
	args := []string{"-Rs", pkg}
	args = append(args, extraArgs...)
	return runInteractive("paru", args...)
}

func (a *AUR) Update(extraArgs ...string) error {
	args := []string{"-Syu"}
	args = append(args, extraArgs...)
	return runInteractive("paru", args...)
}
