package main

import "strings"

type Package struct {
	Name        string
	Version     string
	Description string
	Source      string
	Repo        string
	Installed   bool
	Size        string
}

func (p Package) Format() string {
	var b strings.Builder
	if p.Repo != "" {
		b.WriteString(p.Repo)
		b.WriteString("/")
	}
	b.WriteString(p.Name)
	b.WriteString(" ")
	b.WriteString(p.Version)
	if p.Size != "" {
		b.WriteString(" [")
		b.WriteString(p.Size)
		b.WriteString("]")
	}
	if p.Installed {
		b.WriteString(" [Installed]")
	}
	return b.String()
}

type Manager interface {
	Name() string
	Exists() bool
	Search(query string) ([]Package, error)
	Install(pkg string, extraArgs ...string) error
	Remove(pkg string, extraArgs ...string) error
	Update(extraArgs ...string) error
}
