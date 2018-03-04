package pkg

import (
	"log"

	"github.com/pkg/errors"
	"mcquay.me/pm/db"
)

// Install fetches and installs pkgs from appropriate remotes.
func Install(root string, pkgs []string) error {
	av, err := db.LoadAvailable(root)
	if err != nil {
		return errors.Wrap(err, "loading available db")
	}

	ms, err := av.Installable(pkgs)
	if err != nil {
		return errors.Wrap(err, "checking ability to install")
	}
	for _, m := range ms {
		log.Printf("fake install %v", m)
	}
	return errors.New("NYI")
}
