package main

import (
	"os"
	"os/exec"
)

func runOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func runInteractive(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func runSudoInteractive(name string, args ...string) error {
	sudoArgs := append([]string{name}, args...)
	return runInteractive("sudo", sudoArgs...)
}
