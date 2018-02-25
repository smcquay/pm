package keyring

import (
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"golang.org/x/crypto/openpgp"

	"mcquay.me/fs"
)

func NewKeyPair(root, name, email string) error {
	pgpDir := filepath.Join(root, "var", "lib", "pm", "pgp")
	if !fs.Exists(pgpDir) {
		if err := os.MkdirAll(pgpDir, 0755); err != nil {
			return errors.Wrap(err, "mk pgp dir")
		}
	}

	e, err := openpgp.NewEntity(name, "pm", email, nil)
	if err != nil {
		errors.Wrap(err, "new entity")
	}
	log.Printf("%+v", e)
	return errors.New("NYI")
}
