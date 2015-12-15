package oci

import "testing"

func TestReadSpec(t *testing.T) {
	s, err := ReadConfigFile("./testdata/test.json")
	if err != nil {
		t.Fatal(err)
	}
}
