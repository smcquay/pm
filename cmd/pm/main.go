package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"mcquay.me/fs"
	"mcquay.me/pm/db"
	"mcquay.me/pm/keyring"
	"mcquay.me/pm/pkg"
)

// Version stores the current version, and is updated at build time.
const Version = "dev"

const usage = `pm: simple, cross-platform system package manager

subcommands:
  available  (av)  -- print out all installable packages
  environ    (env) -- print environment information
  install    (in)  -- install packages
  keyring    (key) -- interact with pm's OpenPGP keyring
  ls               -- list installed packages
  package    (pkg) -- create packages
  pull             -- fetch all available packages from all configured remotes
  remote           -- configure remote pmd servers
  version    (v)   -- print version information
`

const keyUsage = `pm keyring: interact with pm's OpenPGP keyring

subcommands:
  create      (c)  --  create a fresh keypair
  export      (e)  --  export a public key to stdout
  import      (i)  --  import a public key from stdin
  ls               --  list configured key info
  rm               --  remove a key from the keyring
  sign        (s)  --  sign a file
  verify      (v)  --  verify a detached signature
`

const pkgUsage = `pm package: generate pm-compatible packages

subcommands:
  create      (c)  --  create a fresh keypair
`

const remoteUsage = `pm remote: configure remote pmd servers

subcommands:
  add         (a)  --  add a URI
  ls               --  list configured remotes
  rm               --  remove a URI
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
		case "ls":
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
		case "rm":
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
			e, err := keyring.FindSecretEntity(root, signID)
			if err != nil {
				fatalf("find secret key: %v\n", err)
			}
			if err := pkg.Create(e, dir); err != nil {
				fatalf("creating package: %v\n", err)
			}
		default:
			fatalf("unknown package subcommand: %q\n\nusage: %v", sub, pkgUsage)
		}
	case "remote":
		if len(os.Args[1:]) < 2 {
			fatalf("pm remote: insufficient args\n\nusage: %v", remoteUsage)
		}
		sub := os.Args[2]
		args := os.Args[3:]

		if err := mkdirs(root); err != nil {
			fatalf("making pm var directories: %v\n", err)
		}

		switch sub {
		case "add", "a":
			if len(args) < 1 {
				fatalf("missing arg\n\nusage: pm remote add [<uris>]\n")
			}
			if err := db.AddRemotes(root, args); err != nil {
				fatalf("remote add: %v\n", err)
			}
		case "rm":
			if len(args) < 1 {
				fatalf("missing arg\n\nusage: pm remote rm [<uris>]\n")
			}
			if err := db.RemoveRemotes(root, args); err != nil {
				fatalf("remote remove: %v\n", err)
			}
		case "ls":
			if err := db.ListRemotes(root, os.Stdout); err != nil {
				fatalf("list: %v\n", err)
			}
		default:
			fatalf("unknown package subcommand: %q\n\nusage: %v", sub, remoteUsage)
		}
	case "pull":
		if err := db.Pull(root); err != nil {
			fatalf("pulling available packages: %v\n", err)
		}
	case "available", "av":
		if err := db.ListAvailable(root, os.Stdout); err != nil {
			fatalf("pulling available packages: %v\n", err)
		}
	case "install", "in":
		if len(os.Args[1:]) < 2 {
			fatalf("pm install: insufficient args\n\nusage: pm install [pkg1, pkg2, ..., pkgN]\n")
		}
		pkgs := os.Args[2:]
		if err := pkg.Install(root, pkgs); err != nil {
			fatalf("installing: %v\n", err)
		}
	case "ls":
		if err := db.ListInstalled(root, os.Stdout); err != nil {
			fatalf("listing installed: %v\n", err)
		}
	case "version", "v":
		fmt.Printf("pm: version %v\n", Version)
	default:
		fatalf("uknown subcommand %q\n\nusage: %v", cmd, usage)
	}
}

func fatalf(f string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, f, args...)
	os.Exit(1)
}

func mkdirs(root string) error {
	d := filepath.Join(root, "var", "lib", "pm")
	if !fs.Exists(d) {
		if err := os.MkdirAll(d, 0700); err != nil {
			return errors.Wrap(err, "mk pm dir")
		}
	}
	return nil
}
