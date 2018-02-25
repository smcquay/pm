package main

import (
	"fmt"
	"os"
)

const usage = `pm: simple, cross-platform system package manager

subcommands:
  keyring    (key) -- interact with pm's OpenPGP keyring
`

const keyUsage = `pm keyring: interact with pm's OpenPGP keyring

subcommands:
  create      (c)  --  create a fresh keypair
`

func main() {
	if len(os.Args) < 2 {
		fatalf("pm: missing subcommand\n\n%v", usage)
	}
	cmd := os.Args[1]

	switch cmd {
	case "key", "keyring":
		if len(os.Args[1:]) < 2 {
			fatalf("pm keyring: insufficient args\n\nusage: %v", keyUsage)
		}
		sub := os.Args[2]
		switch sub {
		case "c", "create":
			fmt.Printf("creating keyring ...\n")
			fatalf("NYI\n")
		default:
			fatalf("unknown keyring subcommand: %q\n\nusage: %v", sub, keyUsage)
		}
	default:
		fatalf("uknown subcommand %q\n\nusage: %v", cmd, usage)
	}
}

func fatalf(f string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, f, args...)
	os.Exit(1)
}
