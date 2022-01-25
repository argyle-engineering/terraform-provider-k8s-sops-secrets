package provider

import (
	"fmt"
	"gopkg.in/yaml.v2"
)

type StringData map[string]string
type Data map[string]string

type secret struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	Type       string     `yaml:"type"`
	StringData StringData `yaml:"stringData"`
	Data       Data       `yaml:"data"`
}

func NewSecret(name string) *secret {
	s := new(secret)
	s.APIVersion = "v1"
	s.Kind = "Secret"
	s.Type = "Opaque"
	s.Metadata.Name = name
	return s
}

func (s secret) Marshall() (string, error) {
	d, err := yaml.Marshal(&s)

	if err != nil {
		err = fmt.Errorf("failed to marshall to yaml: %v", err)
	}

	return string(d), err
}
