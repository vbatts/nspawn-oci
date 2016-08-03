package oci

import (
	"errors"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/vbatts/nspawn-oci/nspawn"
)

func BundleToContainer(bundlepath string) (*Wrapper, error) {
	s, err := ReadConfigFile(filepath.Join(bundlepath, "config.json"))
	if err != nil {
		return nil, err
	}

	if strings.ToLower(s.Platform.OS) != "linux" {
		return nil, errors.New("only linux bundles are supported")
	}

	root, err := filepath.Abs(filepath.Join(bundlepath, s.Root.Path))
	if err != nil {
		return nil, err
	}

	c := nspawn.NewContainer(root)
	c.ReadOnly = s.Root.Readonly
	c.Machine = s.Hostname

	// Personality allows for x86 and x86_64, but this _could_ be any number of
	// other archs ...
	switch s.Platform.Arch {
	case "amd64", "x86_64":
		c.Personality = "x86-64"
	case "x86", "i386", "i586", "i686", "ix86":
		c.Personality = "x86"
	}

	for _, e := range s.Process.Env {
		c.Env = append(c.Env, e)
	}

	// Not sure about handling s.Process.Cwd

	return &Wrapper{
		Spec:      s,
		Container: c,
	}, nil
}

type Wrapper struct {
	Spec      *specs.Spec
	Container *nspawn.Container
}

func (w *Wrapper) Cmd() *exec.Cmd {
	if w.Container == nil {
		return nil
	}

	// TODO check if w.Spec.Mounts is populated, that they have respective src dirs

	return w.Container.Cmd(w.Spec.Process.Args...)
}
