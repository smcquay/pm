package db

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

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

// RemoveInstalled adds m to the installed package database.
func RemoveInstalled(root string, m pm.Meta) error {
	db, err := loadi(root)
	if err != nil {
		return errors.Wrap(err, "loading installed db")
	}
	delete(db, m.Name)
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

// ListInstalledFiles prints the contents of a package.
func ListInstalledFiles(root string, w io.Writer, names []string) error {
	for _, name := range names {
		ok, err := IsInstalled(root, pm.Meta{Name: pm.Name(name)})
		if err != nil {
			return errors.Wrap(err, "is installed")
		}
		if !ok {
			return fmt.Errorf("%v not installed", name)
		}
	}

	for _, name := range names {
		fn := filepath.Join(root, "var", "lib", "pm", "installed", name, "bom.sha256")
		f, err := os.Open(fn)
		if err != nil {
			return errors.Wrapf(err, "opening %v's bom", name)
		}
		bom, err := pm.ParseCS(f)
		if err != nil {
			return errors.Wrapf(err, "parsing %v's bom", name)
		}
		if err := f.Close(); err != nil {
			return errors.Wrapf(err, "closing %v's bom", name)
		}

		ks := []string{}
		for k := range bom {
			ks = append(ks, k)
		}
		sort.Strings(ks)
		for _, k := range ks {
			fmt.Fprintf(w, "%v\n", k)
		}
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
