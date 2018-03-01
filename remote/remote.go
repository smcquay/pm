package remote

import (
	"errors"
	"io"
)

// Add appends the provided uri to the list of configured remotes.
func Add(root string, uri []string) error {
	return errors.New("NYI")
}

// Add removes the given uri from the list of configured remotes.
func Remove(root string, uri []string) error {
	return errors.New("NYI")
}

// List prints all configured remotes to w.
func List(root string, w io.Writer) error {
	return errors.New("NYI")
}
