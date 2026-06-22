package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var managers = []Manager{}

func init() {
	for _, m := range []Manager{
		&Pacman{},
		&AUR{},
		&Flatpak{},
	} {
		if m.Exists() {
			managers = append(managers, m)
		}
	}
}

func printUsage() {
	fmt.Println(`Usage: pkg [command] [arguments]

Commands:
  search  <query>              Search packages across all sources
  install [source/]<package>   Install a package (auto-detects source)
  remove  [source/]<package>   Remove a package (source required if ambiguous)
  update                       Update all packages

If no command is given, runs update (pkg = pkg update).
If command is not recognized, treats it as a search (pkg firefox = pkg search firefox).

Sources: pacman, aur, flatpak
Examples:
  pkg                              # update all
  pkg firefox                      # search
  pkg search firefox               # search
  pkg install firefox              # install (auto-detect)
  pkg install aur/firefox-nightly  # install from AUR
  pkg install firefox --noconfirm  # pass flags through
  pkg install flatpak/firefox --user
  pkg remove flatpak/org.mozilla.firefox
  pkg update`)
}

func main() {
	if len(os.Args) < 2 {
		cmdUpdate()
		return
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "help", "-h", "--help":
		printUsage()
	case "search":
		if len(args) < 1 {
			fmt.Println("Usage: pkg search <query>")
			os.Exit(1)
		}
		cmdSearch(strings.Join(args, " "))
	case "install":
		if len(args) < 1 {
			fmt.Println("Usage: pkg install [source/]<package>")
			os.Exit(1)
		}
		cmdInstall(args[0], args[1:]...)
	case "remove":
		if len(args) < 1 {
			fmt.Println("Usage: pkg remove [source/]<package>")
			os.Exit(1)
		}
		cmdRemove(args[0], args[1:]...)
	case "update":
		cmdUpdate(args...)
	default:
		cmdSearch(strings.Join(os.Args[1:], " "))
	}
}

func cmdSearch(query string) {
	type result struct {
		source string
		pkgs   []Package
		err    error
	}

	ch := make(chan result, len(managers))
	for _, m := range managers {
		m := m
		go func() {
			pkgs, err := m.Search(query)
			ch <- result{m.Name(), pkgs, err}
		}()
	}

	var allPkgs []Package
	for range managers {
		r := <-ch
		if r.err != nil {
			fmt.Fprintf(os.Stderr, "error searching %s: %v\n", r.source, r.err)
			continue
		}
		allPkgs = append(allPkgs, r.pkgs...)
	}

	if len(allPkgs) == 0 {
		fmt.Println("No packages found.")
		return
	}

	for i, j := 0, len(allPkgs)-1; i < j; i, j = i+1, j-1 {
		allPkgs[i], allPkgs[j] = allPkgs[j], allPkgs[i]
	}

	n := len(allPkgs)
	nw := 1
	if n >= 10 {
		nw = 2
	}
	if n >= 100 {
		nw = 3
	}

	for i, pkg := range allPkgs {
		num := n - i
		fmt.Printf("%*d  %s\n", nw, num, pkg.Format())
		if pkg.Description != "" {
			fmt.Printf("%*s  %s\n", nw, "", pkg.Description)
		}
	}

	fmt.Println()
	fmt.Print(":: Packages to install (eg: 1 2 3, 1-3): ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return
	}

	indices := parseSelection(input, n)
	for _, idx := range indices {
		pkg := allPkgs[idx]
		fmt.Printf("Installing %s/%s from %s...\n", pkg.Repo, pkg.Name, pkg.Source)
		if err := installPackage(pkg.Source + "/" + pkg.Name); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
	}
}

func parseSelection(input string, total int) []int {
	parts := strings.Fields(input)
	seen := make(map[int]bool)
	var indices []int

	for _, part := range parts {
		if strings.Contains(part, "-") {
			rangeParts := strings.SplitN(part, "-", 2)
			start, err1 := strconv.Atoi(rangeParts[0])
			end, err2 := strconv.Atoi(rangeParts[1])
			if err1 != nil || err2 != nil || start < 1 || end > total || start > end {
				continue
			}
			for n := start; n <= end; n++ {
				idx := total - n
				if !seen[idx] {
					seen[idx] = true
					indices = append(indices, idx)
				}
			}
		} else {
			n, err := strconv.Atoi(part)
			if err != nil || n < 1 || n > total {
				continue
			}
			idx := total - n
			if !seen[idx] {
				seen[idx] = true
				indices = append(indices, idx)
			}
		}
	}

	return indices
}

func installPackage(name string, extraArgs ...string) error {
	if src, pkg, ok := strings.Cut(name, "/"); ok {
		for _, m := range managers {
			if strings.EqualFold(m.Name(), src) {
				return m.Install(pkg, extraArgs...)
			}
		}
		return fmt.Errorf("unknown source: %s", src)
	}

	for _, m := range managers {
		if err := m.Install(name, extraArgs...); err == nil {
			return nil
		}
	}
	return fmt.Errorf("failed to install %s from any source", name)
}

func cmdInstall(name string, extraArgs ...string) {
	if err := installPackage(name, extraArgs...); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func cmdRemove(name string, extraArgs ...string) {
	if src, pkg, ok := strings.Cut(name, "/"); ok {
		for _, m := range managers {
			if strings.EqualFold(m.Name(), src) {
				fmt.Printf("Removing %s from %s...\n", pkg, m.Name())
				if err := m.Remove(pkg, extraArgs...); err != nil {
					os.Exit(1)
				}
				return
			}
		}
		fmt.Fprintf(os.Stderr, "Unknown source: %s (use: pacman, aur, flatpak)\n", src)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "error: source required for remove (use e.g. pacman/%s or aur/%s)\n", name, name)
	os.Exit(1)
}

func cmdUpdate(extraArgs ...string) {
	skipPacman := false
	for _, m := range managers {
		if m.Name() == "aur" {
			skipPacman = true
			break
		}
	}

	for _, m := range managers {
		if m.Name() == "pacman" && skipPacman {
			continue
		}
		fmt.Printf("\n=== Updating %s ===\n", m.Name())
		if err := m.Update(extraArgs...); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
	}
}
