package pm

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Installed tracks installed packages.
type Installed map[Name]Meta

// Traverse returns a chan of Meta that will be sanely sorted.
func (i Installed) Traverse() <-chan Meta {
	r := make(chan Meta)
	go func() {
		names := Names{}
		for n := range i {
			names = append(names, n)
		}
		sort.Sort(names)

		for _, n := range names {
			r <- i[n]
		}
		close(r)
	}()
	return r
}

// Removable calculates if the packages requested in "in" can all be removed.
func (i Installed) Removable(names []string) (Metas, error) {
	inm := map[Name]bool{}

	// XXX (sm): here we simply check if the package exists; eventually we'll
	// have to check transitive dependencies, and deal with explicitly and
	// implicitly installed packages.

	found := map[Name]Meta{}
	for _, name := range names {
		n := Name(name)
		inm[n] = true
		if m, ok := i[n]; ok {
			found[n] = m
		}
	}

	if len(found) > len(inm) {
		return nil, errors.New("should not have been able to find more than asked for, but did; internals are inconsistent.")
	} else if len(inm) > len(found) {
		// user asked for something that isn't installed.
		missing := []string{}
		for _, name := range names {
			if _, ok := found[Name(name)]; !ok {
				missing = append(missing, name)
			}
		}
		return nil, fmt.Errorf("packages not installed: %v", strings.Join(missing, ", "))
	}

	if len(found) != len(inm) {
		return nil, fmt.Errorf("escapes logic")
	}

	// XXX (sm): the ordering here will also eventually depend on transitive
	// dependencies.
	r := Metas{}
	for _, m := range found {
		r = append(r, m)
	}

	return r, nil
}
