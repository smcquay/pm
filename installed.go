package pm

import "sort"

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
