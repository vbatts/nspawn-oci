package oci

import (
	"encoding/json"
	"io"
	"os"

	"github.com/opencontainers/runtime-spec/specs-go"
)

// ReadConfigFile unmarshals a JSON body from filename and returns an OCI Spec
func ReadConfigFile(filename string) (*specs.Spec, error) {
	fh, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	return ReadConfig(fh)
}

// ReadConfig unmarshals a JSON body from input and returns an OCI Spec
func ReadConfig(input io.Reader) (*specs.Spec, error) {
	d := json.NewDecoder(input)
	s := specs.Spec{}
	if err := d.Decode(&s); err != nil {
		return nil, err
	}
	return &s, nil
}
