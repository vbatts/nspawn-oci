package nspawn

// systemd-nspawn records of when flags were introduced
// https://github.com/systemd/systemd/blob/master/NEWS

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	// DefaultNspawnPath is the standard location of the nspawn binary
	DefaultNspawnPath = "/usr/bin/systemd-nspawn"

	// DefaultContainerDir is where machinectl stores the directories of root
	// filesystems, by default.
	DefaultContainerDir = "/var/lib/machines"
)

// MachinesAvailable returns the list of available system container filesystem paths
func MachinesAvailable() ([]string, error) {
	return filepath.Glob(filepath.Join(DefaultContainerDir, "*"))
}

// NewContainer is a convenience method for framing up container access to the system nspawn
func NewContainer(rootfs string) (*Container, error) {
	npath, err := exec.LookPath(filepath.Base(DefaultNspawnPath))
	if err != nil {
		return nil, err
	}
	n := Nspawn{Path: npath}
	return n.Container(rootfs), nil
}

// Nspawn is a producer of calls to a container
type Nspawn struct {
	Path    string
	version int
}

// Container produces a customizable instance of a contaner to be called using
// this Nspawn executable
func (n *Nspawn) Container(path string) *Container {
	return &Container{
		Nspawn: n,
		Dir:    path,
	}
}

// Version returns the parsed number of this systemd-nspawn version
func (n *Nspawn) Version() (int, error) {
	if n.Path == "" {
		n.Path = DefaultNspawnPath
	}
	if n.version > 0 {
		return n.version, nil
	}
	npath, err := exec.LookPath(filepath.Base(n.Path))
	if err != nil {
		return 0, err
	}
	cmd := exec.Command(npath, "--version")
	outBuf := bytes.NewBuffer(nil)
	errBuf := bytes.NewBuffer(nil)
	cmd.Stdout = outBuf
	cmd.Stderr = errBuf
	err = cmd.Run()
	if err != nil {
		return 0, err
	}
	v := strings.Split(strings.Split(strings.Trim(outBuf.String(), "\n"), "\n")[0], " ")[1]
	i, err := strconv.Atoi(v)
	if err != nil {
		return i, err
	}
	n.version = i
	return i, nil
}

// Container is a customizable instance for constructing a kernel container with systemd-nspawn
//
// for more information see man page systemd-nspawn(1)
type Container struct {
	Nspawn              *Nspawn
	Dir                 string   // directory of the rootfs for this container
	AdditionalArgs      []string // place for flags to systemd-nspawn that are not covered by this struct
	Env                 []string
	Tmpfs               []string
	Template            string
	Cwd                 string
	Machine             string // name of this container (default is root of rootfs directory)
	ReadOnly            bool
	Ephemeral           bool
	Quiet               bool
	Boot                bool
	UUID                string // machine-id
	Personality         string
	SELinuxContext      string
	SELinuxAPIFSContext string
	RegisterMachine     bool
	BindMounts          []BindMountParam // array of PATH[:PATH[:OPTIONS]] (see systemd-nspawn(1) for more info)
	BindRoMounts        []BindMountParam // array of PATH[:PATH[:OPTIONS]] (see systemd-nspawn(1) for more info)
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

func (c *Container) flagMachine() []string {
	if c.Machine != "" {
		return []string{"--machine", c.Machine}
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

func (c *Container) flagSELinuxAPIFSContext() []string {
	if c.SELinuxAPIFSContext != "" {
		return []string{"--selinux-apifs-context", c.SELinuxAPIFSContext}
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

func (c *Container) flagChdir() []string {
	// added in 229
	if i, err := c.Nspawn.Version(); err == nil && i >= 229 {
		return []string{fmt.Sprintf("--chdir=%s", c.Cwd)}
	}
	return []string{}
}

func (c *Container) flagAdditionalArgs() []string {
	return c.AdditionalArgs
}

func (c *Container) flagBindMounts() []string {
	args := []string{}
	for _, bmp := range c.BindMounts {
		args = append(args, fmt.Sprintf("--bind=%s", bmp.String()))
	}
	for _, bmp := range c.BindRoMounts {
		args = append(args, fmt.Sprintf("--bind-ro=%s", bmp.String()))
	}
	return args
}

// this will default to not registering the container with systemd-machined
func (c *Container) flagRegisterMachine() []string {
	if c.RegisterMachine {
		return []string{"--register=true"}
	}
	return []string{"--register=false"}
}

type flagFunc func() []string

func (c *Container) args() []string {
	a := []string{}
	for _, fun := range []flagFunc{
		c.flagBoot,
		c.flagBindMounts,
		c.flagChdir,
		c.flagTmpfs,
		c.flagMachine,
		c.flagTemplate,
		c.flagSELinuxContext,
		c.flagSELinuxAPIFSContext,
		c.flagPersonality,
		c.flagUUID,
		c.flagReadOnly,
		c.flagQuiet,
		c.flagDirectory,
		c.flagSetEnv,
		c.flagEphemeral,
		c.flagAdditionalArgs,
		c.flagRegisterMachine,
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

// BindMountParam for the parameter passed to --bind= and --bind-ro=
type BindMountParam struct {
	Src     string
	Dest    string
	Options string
}

// String renders the fields into PATH[:PATH[:OPTIONS]]
func (bmp BindMountParam) String() string {
	// XXX what to do here? give them a tmp directory if this is empty?
	if bmp.Src == "" {
		return ""
	}
	if bmp.Dest != "" {
		if bmp.Options != "" {
			return fmt.Sprintf("%s:%s:%s", bmp.Src, bmp.Dest, bmp.Options)
		}
		return fmt.Sprintf("%s:%s", bmp.Src, bmp.Dest)
	}
	return bmp.Src
}
