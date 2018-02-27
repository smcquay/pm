package main

import (
	"bufio"
	"fmt"
	"os"

	"mcquay.me/pm/keyring"
	"mcquay.me/pm/pkg"
)

const usage = `pm: simple, cross-platform system package manager

subcommands:
  environ    (env) -- print environment information
  keyring    (key) -- interact with pm's OpenPGP keyring
  package    (pkg) -- create packages
`

const keyUsage = `pm keyring: interact with pm's OpenPGP keyring

subcommands:
  create      (c)  --  create a fresh keypair
  export      (e)  --  export a public key to stdout
  import      (i)  --  import a public key from stdin
  list        (ls) --  list configured key info
  remove      (rm) --  remove a key from the keyring
  sign        (s)  --  sign a file
  verify      (v)  --  verify a detached signature
`

const pkgUsage = `pm package: generate pm-compatible packages

subcommands:
  create      (c)  --  create a fresh keypair
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
	signID := os.Getenv("PM_PGP_ID")

	switch cmd {
	case "env", "environ":
		fmt.Printf("PM_ROOT=%q\n", root)
		fmt.Printf("PM_PGP_ID=%q\n", signID)
	case "key", "keyring":
		if len(os.Args[1:]) < 2 {
			fatalf("pm keyring: insufficient args\n\nusage: %v", keyUsage)
		}
		sub, args := os.Args[2], os.Args[3:]
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
		case "export", "e":
			if len(args) != 1 {
				fatalf("missing email argument\n")
			}
			email := args[0]
			if err := keyring.Export(root, os.Stdout, email); err != nil {
				fatalf("exporting public key for %q: %v\n", email, err)
			}
		case "sign", "s":
			if signID == "" {
				fatalf("must set PM_PGP_ID\n")
			}
			e, err := keyring.FindSecretEntity(root, signID)
			if err != nil {
				fatalf("find secret key: %v\n", err)
			}
			if err := keyring.Sign(e, os.Stdin, os.Stdout); err != nil {
				fatalf("signing: %v\n", err)
			}
		case "verify", "v":
			if len(args) != 2 {
				fatalf("usage: pm key verify <file> <sig>\n")
			}
			fn, sn := args[0], args[1]
			ff, err := os.Open(fn)
			if err != nil {
				fatalf("opening %q: %v\n", fn, err)
			}
			defer ff.Close()
			sf, err := os.Open(sn)
			if err != nil {
				fatalf("opening %q: %v\n", fn, err)
			}
			defer sf.Close()
			if err := keyring.Verify(root, ff, sf); err != nil {
				fatalf("detached sig verify: %v\n", err)
			}
		case "i", "import":
			if err := keyring.Import(root, os.Stdin); err != nil {
				fatalf("importing key: %v\n", err)
			}
		case "remove", "rm":
			if len(args) != 1 {
				fatalf("missing key id\n\nusage: pm key remove <id>\n")
			}
			id := args[0]
			if err := keyring.Remove(root, id); err != nil {
				fatalf("removing key for %q: %v\n", id, err)
			}
		default:
			fatalf("unknown keyring subcommand: %q\n\nusage: %v", sub, keyUsage)
		}
	case "package", "pkg":
		if len(os.Args[1:]) < 2 {
			fatalf("pm package: insufficient args\n\nusage: %v", pkgUsage)
		}
		sub := os.Args[2]
		switch sub {
		case "create", "creat", "c":
			if signID == "" {
				fatalf("must set PM_PGP_ID\n")
			}
			args := os.Args[3:]
			if len(args) != 1 {
				fatalf("usage: pm package create <directory>\n")
			}
			dir := args[0]
			if err := pkg.Create(root, signID, dir); err != nil {
				fatalf("creating package: %v\n", err)
			}
		default:
			fatalf("unknown package subcommand: %q\n\nusage: %v", sub, pkgUsage)
		}
	default:
		fatalf("uknown subcommand %q\n\nusage: %v", cmd, usage)
	}
}

func fatalf(f string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, f, args...)
	os.Exit(1)
}
