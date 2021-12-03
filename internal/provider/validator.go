package provider

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

type Binary interface {
	install() error
}

func Exists(binaryName string) error {
	err, _ := LocalExecutor("command", "-v", binaryName)
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

type Kubectl struct{}

func (k *Kubectl) install() error {

	if err := Exists("kubectl"); err == nil {
		return nil
	}

	var url string

	switch runtime.GOOS {
	case "linux":
		url = "/bin/linux/amd64/kubectl"
	case "darwin":
		if runtime.GOARCH == "amd64" {
			url = "/bin/darwin/amd64/kubectl"
		} else if runtime.GOARCH == "arm64" {
			url = "/bin/darwin/arm64/kubectl"
		} else {
			return fmt.Errorf("unknown darwin arquitecture '%s' is not supported", runtime.GOARCH)
		}
	case "windows":
		return fmt.Errorf("windows is not supported")
	}

	// TODO: better error handling here
	out, _ := os.Create("kubectl")
	defer out.Close()

	resp, _ := http.Get(fmt.Sprintf("https://dl.k8s.io/release/v1.22.4/%s", url))
	defer resp.Body.Close()

	_, _ = io.Copy(out, resp.Body)

	err, _ := LocalExecutor("chmod", "+x", "./kubectl")
	if err != nil {
		return fmt.Errorf("failed to apply permission on kubectl file: %s", err)
	}

	err, _ = LocalExecutor("mv", "./kubectl", "/usr/local/bin/kubectl")
	if err != nil {
		return fmt.Errorf("failed to move kubectl file: %s", err)
	}

	if err = Exists("kubectl"); err != nil {
		return fmt.Errorf("failed to install kubectl")
	}

	return nil
}

type SOPS struct{}

func (s *SOPS) install() error {

	if err := Exists("sops"); err == nil {
		return nil
	}

	var url string

	switch runtime.GOOS {
	case "linux":
		url = "sops-v3.7.1.linux"
	case "darwin":
		if runtime.GOARCH == "amd64" {
			url = "sops-v3.7.1.darwin"
		} else {
			return fmt.Errorf("unknown darwin arquitecture '%s' is not supported", runtime.GOARCH)
		}
	case "windows":
		return fmt.Errorf("windows is not supported")
	}

	// TODO: better error handling here
	out, _ := os.Create("sops")
	defer out.Close()

	resp, _ := http.Get(fmt.Sprintf("https://github.com/mozilla/sops/releases/download/v3.7.1/%s", url))
	defer resp.Body.Close()

	_, _ = io.Copy(out, resp.Body)

	err, _ := LocalExecutor("chmod", "+x", "./sops")
	if err != nil {
		return fmt.Errorf("failed to apply permission on sops file: %s", err)
	}

	err, _ = LocalExecutor("mv", "./sops", "/usr/local/bin/sops")
	if err != nil {
		return fmt.Errorf("failed to move sops file: %s", err)
	}

	if err = Exists("sops"); err != nil {
		return fmt.Errorf("failed to install sops")
	}

	return nil
}
