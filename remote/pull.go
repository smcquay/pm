package remote

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"mcquay.me/pm"
)

const an = "var/lib/pm/available.json"

// Pull updates the available package database.
func Pull(root string) error {
	db, err := load(root)
	if err != nil {
		return errors.Wrap(err, "loading db")
	}

	o := pm.Available{}

	// Order here is important: the guarantee made is that any packages that
	// exist in multiple remotes will be fetched by the first configured
	// remote, which is why we traverse the database in reverse.
	//
	// TODO (sm): make this concurrent
	for i := range db {
		u := db[len(db)-i-1]
		resp, err := http.Get(u.String() + "/available.json")
		if err != nil {
			return errors.Wrap(err, "http get")
		}

		a := pm.Available{}
		if err := json.NewDecoder(resp.Body).Decode(&a); err != nil {
			return errors.Wrap(err, "decode remote available")
		}
		a.SetRemote(u)
		o.Update(a)
	}
	if err := saveAvailable(root, o); err != nil {
		return errors.Wrap(err, "saving available db")
	}
	return nil
}

func saveAvailable(root string, db pm.Available) error {
	f, err := os.Create(filepath.Join(root, an))
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
