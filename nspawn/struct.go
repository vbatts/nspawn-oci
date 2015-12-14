package nspawn

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

var (
	// DefaultNspawn is the default binary called as nspawn
	DefaultNspawn = Nspawn{
		Path: "/usr/bin/systemd-nspawn",
	}

	// DefaultContainerDir is where machinectl stores the directories of root
	// filesystems, by default.
	DefaultContainerDir = "/var/lib/machines"
)

// MachinesAvailable returns the list of available system container filesystem paths
func MachinesAvailable() ([]string, error) {
	return filepath.Glob(filepath.Join(DefaultContainerDir, "*"))
}

// NewContainer is a convenience method for using the DefaultNspawn
func NewContainer(rootfs string) *Container {
	return DefaultNspawn.Container(rootfs)
}

// Nspawn is a producer of calls to a container
type Nspawn struct {
	Path string
}

// Container produces a customizable instance of a contaner to be called using
// this Nspawn executable
func (n *Nspawn) Container(path string) *Container {
	return &Container{
		Nspawn: n,
		Dir:    path,
	}
}

// Container is a customizable instance for constructing a kernel container with systemd-nspawn
//
// for more information see man page systemd-nspawn(1)
type Container struct {
	Nspawn         *Nspawn
	Dir            string
	AdditionalArgs []string
	Env            []string
	Tmpfs          []string
	Template       string
	ReadOnly       bool
	Ephemeral      bool
	Quiet          bool
	Boot           bool
	UUID           string // machine-id
	Personality    string
	SELinuxContext string
}

// flagSetEnv produces the Nspawn flags for setting the needed environment
// variables
func (c *Container) flagSetEnv() []string {
	setEnvFlags := []string{}
	for i := range c.Env {
		setEnvFlags = append(setEnvFlags, fmt.Sprintf("--setenv=%s", c.Env[i]))
	}
	return setEnvFlags
}

func (c *Container) flagTemplate() []string {
	if c.Template != "" {
		return []string{"--template", c.Template}
	}
	return []string{}
}

func (c *Container) flagEphemeral() []string {
	if c.Ephemeral {
		return []string{"--ephemeral"}
	}
	return []string{}
}

func (c *Container) flagReadOnly() []string {
	if c.ReadOnly {
		return []string{"--read-only"}
	}
	return []string{}
}

func (c *Container) flagTmpfs() []string {
	tmpfsFlags := []string{}
	for i := range c.Tmpfs {
		tmpfsFlags = append(tmpfsFlags, fmt.Sprintf("--tmpfs=%s", c.Tmpfs[i]))
	}
	return tmpfsFlags
}

func (c *Container) flagPersonality() []string {
	if c.Personality == "x86" || c.Personality == "x86_64" {
		return []string{"--personality", c.Personality}
	}
	return []string{}
}

func (c *Container) flagSELinuxContext() []string {
	if c.SELinuxContext != "" {
		return []string{"--selinux-context", c.SELinuxContext}
	}
	return []string{}
}

func (c *Container) flagUUID() []string {
	if c.UUID != "" {
		return []string{"--uuid", c.UUID}
	}
	return []string{}
}

func (c *Container) flagBoot() []string {
	if c.Boot {
		return []string{"-b"}
	}
	return []string{}
}

func (c *Container) flagQuiet() []string {
	if c.Quiet {
		return []string{"-q"}
	}
	return []string{}
}

func (c *Container) flagDirectory() []string {
	return []string{"-D", c.Dir}
}

func (c *Container) flagAdditionalArgs() []string {
	return c.AdditionalArgs
}

type flagFunc func() []string

func (c *Container) args() []string {
	a := []string{}
	for _, fun := range []flagFunc{
		c.flagBoot,
		c.flagTmpfs,
		c.flagTemplate,
		c.flagSELinuxContext,
		c.flagPersonality,
		c.flagUUID,
		c.flagReadOnly,
		c.flagQuiet,
		c.flagDirectory,
		c.flagSetEnv,
		c.flagEphemeral,
		c.flagAdditionalArgs,
	} {
		a = append(a, fun()...)
	}
	return a
}

// Cmd assembles the ready-to-call command for this container.
//
// From here, the caller can handle stdin, stderr and stdout as well as the
// return of running the command.
func (c *Container) Cmd(arg ...string) *exec.Cmd {
	args := append(c.args(), arg...)
	return exec.Command(c.Nspawn.Path, args...)
}
