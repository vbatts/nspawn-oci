package oci

import (
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

	root, err := filepath.Abs(filepath.Join(bundlepath, s.Root.Path))
	if err != nil {
		return nil, err
	}

	c, err := nspawn.NewContainer(root)
	if err != nil {
		return nil, err
	}
	c.ReadOnly = s.Root.Readonly
	c.Machine = s.Hostname

	for _, e := range s.Process.Env {
		c.Env = append(c.Env, e)
	}

	if s.Process.Cwd != "" {
		c.Cwd = s.Process.Cwd
	}

	for _, m := range s.Mounts {
		if m.Type == "bind" {
			bmp := nspawn.BindMountParam{
				Src:  m.Source,
				Dest: m.Destination,
			}

			isReadOnly := false
			trimmedOptions := []string{}
			for _, o := range m.Options {
				if o == "ro" {
					isReadOnly = true
				}
				if o != "ro" {
					trimmedOptions = append(trimmedOptions, o)
				}
			}
			if len(trimmedOptions) != 0 {
				bmp.Options = strings.Join(trimmedOptions, ",")
			}

			if isReadOnly {
				c.BindRoMounts = append(c.BindRoMounts, bmp)
			} else {
				c.BindMounts = append(c.BindMounts, bmp)
			}
		}
		// TODO contemplate sysfs, none, devtmpfs, proc, etc.
	}

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
