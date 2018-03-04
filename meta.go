package pm

import (
	"errors"
	"fmt"
	"net/url"
)

// Meta tracks metadata for a package
type Meta struct {
	Name        Name    `json:"name"`
	Version     Version `json:"version"`
	Description string  `json:"description"`

	Remote url.URL `json:"remote"`
}

// Valid validates the contents of a Meta for requires fields.
func (m Meta) Valid() (bool, error) {
	if m.Name == "" {
		return false, errors.New("name cannot be empty")
	}
	if m.Version == "" {
		return false, errors.New("version cannot be empty")
	}
	if m.Description == "" {
		return false, errors.New("description cannot be empty")
	}
	return true, nil
}

// URL returns the http location of this package.
func (m Meta) URL() string {
	return fmt.Sprintf("%s/%s-%s.pkg", m.Remote.String(), m.Name, m.Version)
}

func (m Meta) String() string {
	return m.URL()
}

// Metas is a slice of Meta
type Metas []Meta
