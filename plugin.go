package main

import (
	"encoding/json"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

//go:embed plugins_default.json
var defaultConfig []byte

const userConfigPath = "~/.config/pkgmgr/plugins.json"

type ghPkg struct {
	Repo        string   `json:"repo"`
	Description string   `json:"description"`
	CloneDir    string   `json:"cloneDir"`
	CheckBin    string   `json:"checkBin"`
	Install     []string `json:"install"`
	Update      []string `json:"update"`
}

type pluginConfig struct {
	Github map[string]map[string]ghPkg `json:"github"`
}

type PluginManager struct {
	mu     sync.Mutex
	loaded bool
	config pluginConfig
}

func (pm *PluginManager) Name() string { return "gh" }

func (pm *PluginManager) Exists() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

func (pm *PluginManager) loadConfig() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if pm.loaded {
		return nil
	}

	if err := json.Unmarshal(defaultConfig, &pm.config); err != nil {
		return fmt.Errorf("parsing embedded config: %w", err)
	}

	userPath := expandPath(userConfigPath)
	if data, err := os.ReadFile(userPath); err == nil {
		var userCfg pluginConfig
		if err := json.Unmarshal(data, &userCfg); err != nil {
			return fmt.Errorf("parsing %s: %w", userPath, err)
		}
		for user, pkgs := range userCfg.Github {
			if pm.config.Github == nil {
				pm.config.Github = make(map[string]map[string]ghPkg)
			}
			if _, ok := pm.config.Github[user]; !ok {
				pm.config.Github[user] = make(map[string]ghPkg)
			}
			for name, pkg := range pkgs {
				pm.config.Github[user][name] = pkg
			}
		}
	}

	pm.loaded = true
	return nil
}

func pkgInstalled(pkg ghPkg) bool {
	if _, err := os.Stat(expandPath(pkg.CloneDir)); err == nil {
		return true
	}
	if pkg.CheckBin != "" {
		if _, err := exec.LookPath(pkg.CheckBin); err == nil {
			return true
		}
	}
	return false
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		home, _ := os.UserHomeDir()
		path = home + path[1:]
	}
	path = os.Expand(path, func(s string) string {
		switch s {
		case "XDG_CONFIG_HOME":
			if v := os.Getenv("XDG_CONFIG_HOME"); v != "" {
				return v
			}
			home, _ := os.UserHomeDir()
			return home + "/.config"
		default:
			return os.Getenv(s)
		}
	})
	return path
}

func (pm *PluginManager) Search(query string) ([]Package, error) {
	if err := pm.loadConfig(); err != nil {
		return nil, err
	}
	q := strings.ToLower(query)
	var pkgs []Package
	for user, userPkgs := range pm.config.Github {
		for name, pkg := range userPkgs {
			if !strings.Contains(strings.ToLower(name), q) &&
				!strings.Contains(strings.ToLower(pkg.Description), q) {
				continue
			}
			installed := pkgInstalled(pkg)
			pkgs = append(pkgs, Package{
				Name:        name,
				Version:     "git",
				Description: pkg.Description,
				Source:      "gh",
				Repo:        user,
				Installed:   installed,
			})
		}
	}
	return pkgs, nil
}

func (pm *PluginManager) lookup(pkg string) (user, name string, found bool) {
	if parts := strings.SplitN(pkg, "/", 2); len(parts) == 2 {
		u, n := parts[0], parts[1]
		if um, ok := pm.config.Github[u]; ok {
			if _, ok := um[n]; ok {
				return u, n, true
			}
		}
	}
	for u, pkgs := range pm.config.Github {
		for n := range pkgs {
			if n == pkg {
				return u, n, true
			}
		}
	}
	return "", "", false
}

func (pm *PluginManager) Install(pkg string, extraArgs ...string) error {
	if err := pm.loadConfig(); err != nil {
		return err
	}
	user, name, found := pm.lookup(pkg)
	if !found {
		return fmt.Errorf("unknown plugin package: %s", pkg)
	}
	return pm.installPkg(user, name)
}

func (pm *PluginManager) clonePkg(user, name string) error {
	def := pm.config.Github[user][name]
	cloneDir := expandPath(def.CloneDir)
	repoDir := filepath.Dir(cloneDir)

	if err := os.MkdirAll(repoDir, 0755); err != nil {
		return fmt.Errorf("creating directory %s: %w", repoDir, err)
	}
	repoURL := fmt.Sprintf("https://github.com/%s.git", def.Repo)
	fmt.Printf("Cloning %s into %s...\n", repoURL, cloneDir)
	cmd := exec.Command("git", "clone", repoURL, cloneDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (pm *PluginManager) installPkg(user, name string) error {
	def := pm.config.Github[user][name]
	cloneDir := expandPath(def.CloneDir)

	if _, err := os.Stat(cloneDir); os.IsNotExist(err) {
		if err := pm.clonePkg(user, name); err != nil {
			return fmt.Errorf("cloning repo: %w", err)
		}
	} else {
		fmt.Printf("Updating existing repo at %s...\n", cloneDir)
		cmd := exec.Command("git", "-C", cloneDir, "pull", "--rebase")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run()
	}

	return pm.runSteps(cloneDir, def.Install)
}

func (pm *PluginManager) Remove(pkg string, extraArgs ...string) error {
	return fmt.Errorf("gh packages cannot be removed automatically; remove manually")
}

func (pm *PluginManager) Update(extraArgs ...string) error {
	if err := pm.loadConfig(); err != nil {
		return err
	}
	for user, userPkgs := range pm.config.Github {
		for name := range userPkgs {
			def := pm.config.Github[user][name]
			cloneDir := expandPath(def.CloneDir)

			if _, err := os.Stat(cloneDir); os.IsNotExist(err) {
				if !pkgInstalled(def) {
					fmt.Printf("Skipping %s/%s (not cloned yet, use install first)\n", user, name)
					continue
				}
				if err := pm.clonePkg(user, name); err != nil {
					fmt.Fprintf(os.Stderr, "error cloning %s: %v\n", name, err)
					continue
				}
			}

			fmt.Printf("\n=== Updating %s/%s ===\n", user, name)

			cmd := exec.Command("git", "-C", cloneDir, "pull", "--rebase")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "error pulling %s: %v\n", name, err)
				continue
			}

			if err := pm.runSteps(cloneDir, def.Update); err != nil {
				fmt.Fprintf(os.Stderr, "error updating %s: %v\n", name, err)
			}
		}
	}
	return nil
}

func (pm *PluginManager) runSteps(dir string, steps []string) error {
	for _, step := range steps {
		fmt.Printf("  => %s\n", step)
		cmd := exec.Command("sh", "-c", step)
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("step %q failed: %w", step, err)
		}
	}
	return nil
}
