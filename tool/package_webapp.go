package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	println("Building polymer webapp.")
	c := exec.Command("polymer", "build", "--add-service-worker")
	c.Dir = "../webapp"
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err := c.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		return
	}

	println("Removing old build.")
	err = os.RemoveAll("../service/src/github.com/rltoscano/pluot/web")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		return
	}

	println("Copying new build.")
	c = exec.Command("powershell", "copy", "../webapp/build/default", "../service/src/github.com/rltoscano/pluot/web", "-recurse")
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err = c.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		return
	}
}
