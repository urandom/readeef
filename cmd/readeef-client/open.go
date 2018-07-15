package main

import (
	"fmt"
	"os/exec"
)

func open(u string) error {
	args := []string{}
	if len(OPEN) > 1 {
		args = append(args, OPEN...)
	}
	args = append(args, u)
	if err := exec.Command(OPEN[0], args...).Run(); err != nil {
		return fmt.Errorf("Error opening %s: %v", u, err)
	}

	return nil
}
