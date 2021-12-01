package bin_man

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

type Binary interface {
	install() error
}

func Exists(binaryName string) error {
	return localExecutor("command", "-v", binaryName)
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

	err := localExecutor("chmod", "+x", "./kubectl")
	if err != nil {
		return fmt.Errorf("failed to apply permission on kubectl file: %s", err)
	}

	err = localExecutor("mv", "./kubectl", "/usr/local/bin/kubectl")
	if err != nil {
		return fmt.Errorf("failed to move kubectl file: %s", err)
	}

	if err = Exists("kubectl"); err != nil {
		return fmt.Errorf("failed to install kubectl")
	}

	return nil
}

func localExecutor(name string, args ...string) error {
	var out bytes.Buffer

	c := exec.Command(name, args...)

	c.Stdout = &out

	err := c.Run()

	if err != nil {
		return err
	}

	return nil
}
