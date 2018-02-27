package pkg

import (
	"fmt"

	"mcquay.me/fs"
)

// Create traverses the contents of dir and emits a valid pkg, signed by id
func Create(dir, id string) error {
	if !fs.Exists(dir) {
		return fmt.Errorf("%q: doesn't exist", dir)
	}
	if !fs.IsDir(dir) {
		return fmt.Errorf("%q: is not a directory", dir)
	}
	return fmt.Errorf("creating package from %q for %q: NYI", dir, id)
}
