package db

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

// TODO (sm): add more tests, including
// - empty add
// - removing db to empty
// - bad uris

func dirMe(t *testing.T) (string, func()) {
	root, err := ioutil.TempDir("", "pm-tests-")
	if err != nil {
		t.Fatalf("tmpdir: %v", err)
	}
	if err := mkdirs(root); err != nil {
		t.Fatalf("making pm dirs: %v", err)
	}
	return root, func() {
		if err := os.RemoveAll(root); err != nil {
			t.Fatalf("cleanup: %v", err)
		}
	}
}

func TestAdd(t *testing.T) {
	root, del := dirMe(t)
	defer del()

	{
		db, err := load(root)
		if err != nil {
			t.Fatalf("load: %v", err)
		}
		if got, want := len(db), 0; got != want {
			t.Fatalf("empty db not empty: got %v, want %v", got, want)
		}
	}

	bad := []string{
		"http\ns://\nFoo|n",
	}

	if err := AddRemotes(root, bad); err == nil {
		t.Fatalf("didn't detect bad url")
	}

	uris := []string{
		"https://pm.mcquay.me/darwin/amd64",
	}

	if err := AddRemotes(root, uris); err != nil {
		t.Fatalf("add: %v", err)
	}

	db, err := load(root)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if got, want := len(db), len(uris); got != want {
		t.Fatalf("unepected number of uris; got %v, want %v", got, want)
	}

	for _, u := range uris {
		found := false
		for _, d := range db {
			if d.String() == u {
				found = true
			}
		}
		if !found {
			t.Fatalf("did not find %v in the db", u)
		}
	}

	if err := Add(root, uris); err == nil {
		t.Fatalf("did not detect duplicate, and should have")
	}
}

func TestRemove(t *testing.T) {
	root, del := dirMe(t)
	defer del()

	if err := Remove(root, nil); err == nil {
		t.Fatalf("should have returned error on empty db")
	}

	uris := []string{
		"https://pm.mcquay.me/foo",
		"https://pm.mcquay.me/bar",
		"https://pm.mcquay.me/baz",
	}

	if err := Remove(root, uris); err == nil {
		t.Fatalf("should have returned error asking to remove many uri on empty db")
	}

	if err := Add(root, uris); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := Remove(root, uris[1:2]); err != nil {
		t.Fatalf("remove: %v", err)
	}
	db, err := load(root)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got, want := len(db), len(uris)-1; got != want {
		t.Fatalf("unepected number of uris; got %v, want %v", got, want)
	}

	found := false
	for _, d := range db {
		if d.String() == uris[1] {
			found = true
		}
	}
	if found {
		for _, v := range db {
			t.Logf("%v", v.String())
		}
		t.Fatalf("failed to remove %v", uris[1:2])
	}
}

func TestList(t *testing.T) {
	root, del := dirMe(t)
	defer del()
	uris := []string{
		"https://pm.mcquay.me/foo",
		"https://pm.mcquay.me/bar",
		"https://pm.mcquay.me/baz",
	}

	if err := Add(root, uris); err != nil {
		t.Fatalf("add: %v", err)
	}

	buf := &bytes.Buffer{}
	if err := List(root, buf); err != nil {
		t.Fatalf("list: %v", err)
	}

	for _, u := range uris {
		if !strings.Contains(buf.String(), u) {
			t.Fatalf("could not find %q in output\n%v", u, buf.String())
		}
	}

}
