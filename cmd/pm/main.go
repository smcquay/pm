package main

import (
	"bufio"
	"fmt"
	"os"

	"mcquay.me/pm/keyring"
)

const usage = `pm: simple, cross-platform system package manager

subcommands:
  environ    (env) -- print environment information
  keyring    (key) -- interact with pm's OpenPGP keyring
`

const keyUsage = `pm keyring: interact with pm's OpenPGP keyring

subcommands:
  create      (c)  --  create a fresh keypair
  list        (ls) --  list configured key info
`

func main() {
	if len(os.Args) < 2 {
		fatalf("pm: missing subcommand\n\n%v", usage)
	}
	cmd := os.Args[1]

	root := os.Getenv("PM_ROOT")
	if root == "" {
		root = "/usr/local"
	}

	switch cmd {
	case "env", "environ":
		fmt.Printf("PM_ROOT=%q\n", root)
	case "key", "keyring":
		if len(os.Args[1:]) < 2 {
			fatalf("pm keyring: insufficient args\n\nusage: %v", keyUsage)
		}
		sub := os.Args[2]
		switch sub {
		case "ls", "list":
			if err := keyring.ListKeys(root, os.Stdout); err != nil {
				fatalf("listing keypair: %v\n", err)
			}
		case "c", "create":
			var name, email string
			s := bufio.NewScanner(os.Stdin)

			fmt.Printf("name: ")
			s.Scan()
			if err := s.Err(); err != nil {
				fatalf("reading name: %v\n", err)
			}
			name = s.Text()

			fmt.Printf("email: ")
			s.Scan()
			if err := s.Err(); err != nil {
				fatalf("reading email: %v\n", err)
			}
			email = s.Text()

			if err := os.Stdin.Close(); err != nil {
				fatalf("%v\n", err)
			}

			if err := keyring.NewKeyPair(root, name, email); err != nil {
				fatalf("creating keypair: %v\n", err)
			}
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
