package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/vbatts/nspawn-oci/oci"
)

var exitCode = 0

func main() {
	defer os.Exit(exitCode) // put this in first, so it is popped last (allowing other defers to happen)
	flag.Parse()

	if flag.NArg() != 1 {
		log.Println("Provide only the path to the OpenContainer bundle.")
		exitCode = 1
		return
	}

	w, err := oci.BundleToContainer(flag.Args()[0])
	if err != nil {
		log.Println(err)
		exitCode = 1
		return
	}

	cmd := w.Cmd()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("[DEBUG] %#v", w.Container)
		log.Printf("[DEBUG] %q (%#v)", strings.Join(cmd.Args, " "), cmd)
		log.Println(err)
		exitCode = 1
		return
	}
}
