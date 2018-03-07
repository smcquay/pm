package pkg

import (
	"log"

	"github.com/pkg/errors"
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
	log.Printf("%+v", ms)

	return errors.New("NYI")
}
