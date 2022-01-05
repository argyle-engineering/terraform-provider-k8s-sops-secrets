package provider

import (
	"fmt"
	"os"
	"testing"
)

func TestMarshall(t *testing.T) {
	s := NewSecret("example")

	dat, _ := os.ReadFile("certificate.pem")

	sd := StringData{
		"certificate": string(dat),
	}
	s.StringData = sd
	marshall, err := s.Marshall()
	if err != nil {
		t.Errorf("tried to marshall and failed: %s", err)
	}

	fmt.Printf("---  dump:\n%s\n\n", marshall)
}
