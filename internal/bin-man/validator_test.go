package bin_man

import (
	"testing"
)

func TestExists(t *testing.T) {
	err := Exists("bash")
	if err != nil {
		t.Errorf("tried to find bash and failed: %s", err)
	}
}

func TestNotExists(t *testing.T) {
	err := Exists("bash_bunny_bugaloo")
	if err == nil {
		t.Errorf("tried to find non-existent-binary and did not fail")
	}
}

func TestKubectlInstall(t *testing.T) {
	kubectl := Kubectl{}
	err := kubectl.install()
	if err != nil {
		t.Errorf("tried to find install kubectl and failed: %s", err)
	}
}
