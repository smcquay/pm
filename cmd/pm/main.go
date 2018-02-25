package main

import (
	"fmt"
	"os"
)

const usage = `pm: simple, cross-platform system package manager

subcommands:
  keyring    (key) -- interact with pm's OpenPGP keyring
`

func main() {
	if len(os.Args) < 2 {
		fatal(usage)
	}
	cmd := os.Args[1]

	switch cmd {
	case "key", "keyring":
	default:
		fatal("uknown subcommand %q\n\nusage: %v", cmd, usage)
	}
}

func fatal(f string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, f, args...)
	os.Exit(1)
}
