package remote

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"mcquay.me/fs"
)

// DB is a slice of available URI
type DB []url.URL

const fn = "var/lib/pm/available.json"

// Add appends the provided uri to the list of configured remotes.
func Add(root string, uris []string) error {
	db, err := load(root)
	if err != nil {
		return errors.Wrap(err, "loading")
	}

	dbm := map[string]bool{}
	for _, u := range db {
		dbm[u.String()] = true
	}

	for _, uri := range uris {
		pu, err := url.Parse(uri)
		if err != nil {
			return errors.Wrap(err, "url parse")
		}

		u := strip(*pu)

		if _, ok := dbm[u.String()]; ok {
			return fmt.Errorf("%q already in db", u.String())
		}
		db = append(db, u)
	}

	return save(root, db)
}

// Remove removes the given uri from the list of configured remotes.
func Remove(root string, uris []string) error {
	db, err := load(root)
	if err != nil {
		return errors.Wrap(err, "loading")
	}

	rms := map[string]bool{}
	for _, uri := range uris {
		pu, err := url.Parse(uri)
		if err != nil {
			return errors.Wrap(err, "url parse")
		}

		u := strip(*pu)

		rms[u.String()] = true
	}

	o := DB{}
	for _, d := range db {
		if _, ok := rms[d.String()]; !ok {
			o = append(o, d)
		}
	}

	if len(o) == len(db) {
		return errors.New("found no matching remotes")
	}

	return save(root, o)
}

// List prints all configured remotes to w.
func List(root string, w io.Writer) error {
	db, err := load(root)
	if err != nil {
		return errors.Wrap(err, "loading")
	}
	for _, u := range db {
		fmt.Fprintf(w, "%s\n", u.String())
	}
	return nil
}

func load(root string) (DB, error) {
	r := DB{}
	dbn := filepath.Join(root, fn)

	if !fs.Exists(dbn) {
		return r, nil
	}

	f, err := os.Open(filepath.Join(root, fn))
	if err != nil {
		return r, errors.Wrap(err, "open")
	}

	if err := json.NewDecoder(f).Decode(&r); err != nil {
		return r, errors.Wrap(err, "decoding db")
	}

	return r, nil
}

func save(root string, db DB) error {
	f, err := os.Create(filepath.Join(root, fn))
	if err != nil {
		return errors.Wrap(err, "create")
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "\t")

	if err := enc.Encode(&db); err != nil {
		return errors.Wrap(err, "decoding db")
	}

	if err := f.Close(); err != nil {
		return errors.Wrap(err, "close db")
	}

	return nil
}

// strip removes all fields we don't currently need.
func strip(u url.URL) url.URL {
	return url.URL{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   u.Path,
	}
}

func mkdirs(root string) error {
	d, _ := filepath.Split(filepath.Join(root, fn))
	if !fs.Exists(d) {
		if err := os.MkdirAll(d, 0700); err != nil {
			return errors.Wrap(err, "mk pm dir")
		}
	}
	return nil
}
