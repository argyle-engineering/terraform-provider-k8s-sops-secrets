package provider

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
)

func Exists(binaryName string) error {
	err, _ := LocalExecutor("which", binaryName)
	return err
}

func LocalExecutor(name string, args ...string) (error, bytes.Buffer) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	c := exec.Command(name, args...)

	c.Stdout = &out
	c.Stderr = &errOut

	err := c.Run()

	if err != nil {
		log.Println(errOut.String())
		return err, out
	}

	return nil, out
}

func ExecuteBash(cmd string, dir string) (string, error) {
	c := exec.Command("bash", "-c", cmd)
	c.Dir = dir
	out, err := c.Output()
	if err != nil {
		return string(out), fmt.Errorf("failed to execute command: %s - %s", cmd, err)
	}
	return string(out), nil
}
