package pm

import (
	"github.com/pkg/errors"
)

// Name exists to document the keys in Available
type Name string

// Version exists to document the keys in Available
type Version string

// Available is the structure used to represent the collection of all packages
// that can be installed.
type Available map[Name]map[Version]Meta

// Add inserts m into a.
func (a Available) Add(m Meta) error {
	if _, err := m.Valid(); err != nil {
		return errors.Wrap(err, "invalid meta")
	}

	if _, ok := a[Name(m.Name)]; !ok {
		a[m.Name] = map[Version]Meta{}
	}
	a[m.Name][m.Version] = m
	return nil
}

// Update inserts all data from o into a.
func (a Available) Update(o Available) error {
	for _, vers := range o {
		for _, m := range vers {
			if err := a.Add(m); err != nil {
				return errors.Wrap(err, "adding")
			}
		}
	}
	return nil
}
