package pm

import (
	"net/url"

	"github.com/pkg/errors"
)

// Name exists to document the keys in Available
type Name string

type Names []Name

func (n Names) Len() int           { return len(n) }
func (n Names) Swap(a, b int)      { n[a], n[b] = n[b], n[a] }
func (n Names) Less(a, b int) bool { return n[a] < n[b] }

// Version exists to document the keys in Available
type Version string
type Versions []Version

// TODO (sm): make this semver sort?
func (n Versions) Len() int           { return len(n) }
func (n Versions) Swap(a, b int)      { n[a], n[b] = n[b], n[a] }
func (n Versions) Less(a, b int) bool { return n[a] < n[b] }

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

func (a Available) SetRemote(u url.URL) {
	for n, vers := range a {
		for v, _ := range vers {
			m := a[n][v]
			m.Remote = u
			a[n][v] = m
		}
	}
}
