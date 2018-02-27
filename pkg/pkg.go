package pkg

import (
	"fmt"
	"log"

	"github.com/pkg/errors"
	"mcquay.me/fs"
	"mcquay.me/pm/keyring"
)

// Create traverses the contents of dir and emits a valid pkg, signed by id
func Create(root, id, dir string) error {
	if !fs.Exists(dir) {
		return fmt.Errorf("%q: doesn't exist", dir)
	}
	if !fs.IsDir(dir) {
		return fmt.Errorf("%q: is not a directory", dir)
	}
	e, err := keyring.FindSecretEntity(dir, id)
	if err != nil {
		return errors.Wrap(err, "find secret key")
	}
	log.Printf("found key: %v", e.PrimaryKey.KeyIdShortString())
	return fmt.Errorf("creating package from %q for %q: NYI", dir, id)
}
