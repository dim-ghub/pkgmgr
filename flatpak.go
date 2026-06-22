package main

import (
	"os/exec"
	"strings"
)

type Flatpak struct{}

func (f *Flatpak) Name() string { return "flatpak" }

func (f *Flatpak) Exists() bool {
	_, err := exec.LookPath("flatpak")
	return err == nil
}

func (f *Flatpak) Search(query string) ([]Package, error) {
	out, err := runOutput("flatpak", "search", query)
	if err != nil {
		return nil, err
	}
	return parseFlatpakOutput(out), nil
}

func parseFlatpakOutput(out string) []Package {
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) < 2 {
		return nil
	}
	lines = lines[1:]

	var pkgs []Package
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 3 {
			continue
		}

		desc := strings.TrimSpace(fields[1])
		id := strings.TrimSpace(fields[2])
		version := ""
		if len(fields) > 3 {
			version = strings.TrimSpace(fields[3])
		}
		pkgs = append(pkgs, Package{
			Name:        id,
			Version:     version,
			Description: desc,
			Source:      "flatpak",
			Repo:        "flatpak",
		})
	}
	return pkgs
}

func (f *Flatpak) Install(pkg string, extraArgs ...string) error {
	args := []string{"install", pkg}
	args = append(args, extraArgs...)
	return runSudoInteractive("flatpak", args...)
}

func (f *Flatpak) Remove(pkg string, extraArgs ...string) error {
	args := []string{"uninstall", pkg}
	args = append(args, extraArgs...)
	return runSudoInteractive("flatpak", args...)
}

func (f *Flatpak) Update(extraArgs ...string) error {
	args := []string{"update"}
	args = append(args, extraArgs...)
	return runSudoInteractive("flatpak", args...)
}
