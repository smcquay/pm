package pm

import (
	"errors"
	"sort"
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
	return nil, errors.New("NYI")
}
