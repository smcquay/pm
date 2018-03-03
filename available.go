package pm

import (
	"net/url"
	"sort"

	"github.com/pkg/errors"
)

// Name exists to document the keys in Available
type Name string

// Names is a slice of names ... with sorting!
type Names []Name

func (n Names) Len() int           { return len(n) }
func (n Names) Swap(a, b int)      { n[a], n[b] = n[b], n[a] }
func (n Names) Less(a, b int) bool { return n[a] < n[b] }

// Version exists to document the keys in Available
type Version string

// Versions is a slice of Version ... with sorting!
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

// SetRemote adds the information in the url to the database.
func (a Available) SetRemote(u url.URL) {
	for n, vers := range a {
		for v := range vers {
			m := a[n][v]
			m.Remote = u
			a[n][v] = m
		}
	}
}

// Traverse returns a chan of Meta that will be sanely sorted.
func (a Available) Traverse() <-chan Meta {
	r := make(chan Meta)
	go func() {
		names := Names{}
		nvs := map[Name]Versions{}
		for n, vers := range a {
			names = append(names, n)
			for v := range vers {
				nvs[n] = append(nvs[n], v)
			}
			sort.Sort(nvs[n])
		}
		sort.Sort(names)

		for _, n := range names {
			for _, v := range nvs[n] {
				r <- a[n][v]
			}
		}
		close(r)
	}()
	return r
}
