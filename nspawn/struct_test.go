package nspawn

import (
	"strings"
	"testing"
)

func TestMachines(t *testing.T) {
	m, err := MachinesAvailable()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("found %q", m)
}

func TestSetEnv(t *testing.T) {
	e := "FARTS=true"

	n := DefaultNspawn
	c := n.Container("/home/containers/debian-jessie/")
	c.Env = append(c.Env, e)
	c.Quiet = true
	cmd := c.Cmd("/bin/cat", "/proc/self/environ")

	t.Logf("Path %q", cmd.Path)
	t.Logf("Args %q \n (%q)", cmd.Args, strings.Join(cmd.Args, " "))
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Error(err)
	}
	found := false
	for _, environ := range strings.Split(string(output), "\x00") {
		if environ == e {
			found = true
		}
	}
	if !found {
		t.Errorf("expected to find %q; got %q", e, string(output))
	}
}
