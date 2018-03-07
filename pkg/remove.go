package pkg

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"mcquay.me/pm"
	"mcquay.me/pm/db"
)

// Remove uninstalls packages.
func Remove(root string, pkgs []string) error {
	iDB, err := db.LoadInstalled(root)
	if err != nil {
		return errors.Wrap(err, "loading available db")
	}

	ms, err := iDB.Removable(pkgs)
	if err != nil {
		return errors.Wrap(err, "checking ability to remove")
	}

	for _, m := range ms {
		if err := script(root, m, "pre-remove"); err != nil {
			return errors.Wrap(err, "pre-remove")
		}

		mdir := filepath.Join(root, installed, string(m.Name))
		bom := filepath.Join(mdir, "bom.sha256")
		bf, err := os.Open(bom)
		if err != nil {
			return errors.Wrapf(err, "%q: opening bom", m.Name)
		}

		cs, err := pm.ParseCS(bf)
		if err != nil {
			return errors.Wrapf(err, "%q: parsing bom", m.Name)
		}

		for n := range cs {
			if err := os.Remove(filepath.Join(root, n)); err != nil {
				return errors.Wrapf(err, "pkg %q", m.Name)
			}
		}

		if err := script(root, m, "post-remove"); err != nil {
			return errors.Wrap(err, "post-remove")
		}

		if err := db.RemoveInstalled(root, m); err != nil {
			return errors.Wrapf(err, "removing %q", m.Name)
		}

		if err := os.RemoveAll(mdir); err != nil {
			return errors.Wrapf(err, "%q: removing pm install dir", m.Name)
		}
	}

	return nil
}
