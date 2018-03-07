package db

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"mcquay.me/fs"
	"mcquay.me/pm"
)

const in = "var/lib/pm/installed.json"

// AddInstalled adds m to the installed package database.
func AddInstalled(root string, m pm.Meta) error {
	db, err := loadi(root)
	if err != nil {
		return errors.Wrap(err, "loading installed db")
	}
	db[m.Name] = m
	return savei(root, db)
}

// IsInstalled checks if m is in the installed package database.
func IsInstalled(root string, m pm.Meta) (bool, error) {
	db, err := loadi(root)
	if err != nil {
		return false, errors.Wrap(err, "loading installed db")
	}

	_, r := db[m.Name]
	return r, nil
}

// ListInstalled pretty prints the installed database to w.
func ListInstalled(root string, w io.Writer) error {
	db, err := loadi(root)
	if err != nil {
		return errors.Wrap(err, "loading installed db")
	}

	for m := range db.Traverse() {
		fmt.Fprintf(w, "%v\t%v\t%v\n", m.Name, m.Version, m.Remote.String())
	}
	return nil
}

func LoadInstalled(root string) (pm.Installed, error) {
	return loadi(root)
}

func loadi(root string) (pm.Installed, error) {
	r := pm.Installed{}
	dbn := filepath.Join(root, in)

	if !fs.Exists(dbn) {
		return r, nil
	}

	f, err := os.Open(dbn)
	if err != nil {
		return r, errors.Wrap(err, "open")
	}

	if err := json.NewDecoder(f).Decode(&r); err != nil {
		return r, errors.Wrap(err, "decoding db")
	}

	return r, nil
}

func savei(root string, db pm.Installed) error {
	f, err := os.Create(filepath.Join(root, in))
	if err != nil {
		return errors.Wrap(err, "create")
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "\t")
	if err := enc.Encode(&db); err != nil {
		return errors.Wrap(err, "decoding db")
	}
	if err := f.Close(); err != nil {
		return errors.Wrap(err, "close db")
	}
	return nil
}
