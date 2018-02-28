package pkg

import (
	"fmt"

	"golang.org/x/crypto/openpgp"
	"mcquay.me/fs"
)

// Create traverses the contents of dir and emits a valid pkg, signed by id
func Create(e *openpgp.Entity, dir string) error {
	if !fs.Exists(dir) {
		return fmt.Errorf("%q: doesn't exist", dir)
	}
	if !fs.IsDir(dir) {
		return fmt.Errorf("%q: is not a directory", dir)
	}
	return fmt.Errorf("creating package from %q for %q: NYI", dir, e.PrimaryKey.KeyIdShortString())
}
