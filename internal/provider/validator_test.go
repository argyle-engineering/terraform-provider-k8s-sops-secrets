package provider

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
