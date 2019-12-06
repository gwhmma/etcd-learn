package main

import (
	"fmt"
	"os/exec"
)

func main() {
	err := exec.Command("/bin/bash", "-c", "echo hello world").Run()
	if err != nil {
		fmt.Println(err)
	}
}
