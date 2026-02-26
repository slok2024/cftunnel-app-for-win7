//go:build windows

package main

import (
	"fmt"
	"os/exec"
)

func quickProcessKill(pid int) error {
	return exec.Command("taskkill", "/PID", fmt.Sprintf("%d", pid)).Run()
}

func quickProcessAlive(pid int) bool {
	err := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/NH").Run()
	return err == nil
}
