package pm

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

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
func (v Versions) Len() int           { return len(v) }
func (v Versions) Swap(a, b int)      { v[a], v[b] = v[b], v[a] }
func (v Versions) Less(a, b int) bool { return v[a] < v[b] }

type label struct {
	n Name
	v Version
}

type labels []label

// TODO (sm): make this semver sort?
func (n labels) Len() int      { return len(n) }
func (n labels) Swap(a, b int) { n[a], n[b] = n[b], n[a] }
func (n labels) Less(a, b int) bool {
	if n[a].n != n[b].n {
		return n[a].n < n[b].n
	}
	return n[a].v < n[b].v
}

// Available is the structure used to represent the collection of all packages
// that can be installed.
type Available map[Name]map[Version]Meta

// Get returns the meta stored at a[n][v] or an error explaining why it could
// not be Get.
func (a Available) Get(n Name, v Version) (Meta, error) {
	if _, ok := a[n]; !ok {
		return Meta{}, errors.Errorf("could not find package named %q", n)
	}

	if v == "" {
		if len(a[n]) == 0 {
			return Meta{}, errors.Errorf("no configured versions for %q", n)
		}
		vers := Versions{}
		for ver := range a[n] {
			vers = append(vers, ver)
		}
		sort.Sort(vers)
		v = vers[len(vers)-1]
	}

	if _, ok := a[n][v]; !ok {
		return Meta{}, errors.Errorf("could not find %v@%v in database", n, v)
	}
	return a[n][v], nil
}

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

func labelForString(s string) (label, error) {
	r := label{}
	c := strings.Count(s, "@")
	switch c {
	case 0:
		r.n = Name(s)
	case 1:
		sp := strings.Split(s, "@")
		r.n, r.v = Name(sp[0]), Version(sp[1])
	default:
		return r, fmt.Errorf("unexpected number of '@' found, got %v, want 1", c)
	}
	if r.n == "" {
		return r, fmt.Errorf("name cannot be empty")
	}
	return r, nil
}

// Installable calculates if the packages requested in "in" can be installed.
func (a Available) Installable(in []string) (Metas, error) {
	ls := labels{}
	for _, i := range in {
		l, err := labelForString(i)
		if err != nil {
			return nil, errors.Wrap(err, "parsing name/version")
		}
		ls = append(ls, l)
	}

	seen := map[Name]bool{}
	for _, l := range ls {
		if _, ok := seen[l.n]; ok {
			return nil, fmt.Errorf("can only ask to install %q once", l.n)
		}
		seen[l.n] = true
	}

	ms := Metas{}
	for _, l := range ls {
		m, err := a.Get(l.n, l.v)
		if err != nil {
			return ms, errors.Wrapf(err, "getting %v", l)
		}
		ms = append(ms, m)
	}

	return ms, nil
}
