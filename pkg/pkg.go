package pkg

import "fmt"

// Create traverses the contents of dir and emits a valid pkg, signed by id
func Create(dir, id string) error {
	return fmt.Errorf("creating package from %q for %q: NYI", dir, id)
}
