package utils

import (
	"fmt"
	"os/exec"
)

func CreateResizedImage(src string, dest string, width, height int) error {
	args := []string{
		"convert",
		src,
		"-resize", fmt.Sprintf("%dx%d^", width, height),
		"-gravity", "Center",
		"-extent", fmt.Sprintf("%dx%d", width, height),
		dest,
	}

	fmt.Printf("args: %v\n", args)

	cmd := exec.Command("magick", args...)
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
