package pm

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// ParseCS returns a parsed checksum file.
func ParseCS(f io.Reader) (map[string]string, error) {
	cs := map[string]string{}
	s := bufio.NewScanner(f)
	for s.Scan() {
		elems := strings.Split(s.Text(), "\t")
		if len(elems) != 2 {
			return nil, fmt.Errorf("manifest format error; got %d elements, want 2", len(elems))
		}
		cs[elems[1]] = elems[0]
	}
	return cs, nil
}
