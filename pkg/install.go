package pkg

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"mcquay.me/fs"
	"mcquay.me/pm"
	"mcquay.me/pm/db"
)

const cache = "var/cache/pm"

// Install fetches and installs pkgs from appropriate remotes.
func Install(root string, pkgs []string) error {
	av, err := db.LoadAvailable(root)
	if err != nil {
		return errors.Wrap(err, "loading available db")
	}

	ms, err := av.Installable(pkgs)
	if err != nil {
		return errors.Wrap(err, "checking ability to install")
	}

	cacheDir := filepath.Join(root, cache)
	if !fs.Exists(cacheDir) {
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return errors.Wrap(err, "creating non-existent cache dir")
		}
	}
	if !fs.IsDir(cacheDir) {
		return errors.Errorf("%q is not a directory!", cacheDir)
	}

	if err := download(cacheDir, ms); err != nil {
		return errors.Wrap(err, "downloading")
	}

	return errors.New("NYI")
}

func download(cache string, ms pm.Metas) error {
	// TODO (sm): concurrently fetch
	for _, m := range ms {
		resp, err := http.Get(m.URL())
		if err != nil {
			return errors.Wrap(err, "http get")
		}
		fn := filepath.Join(cache, m.Pkg())
		f, err := os.Create(fn)
		if err != nil {
			return errors.Wrap(err, "creating")
		}

		if n, err := io.Copy(f, resp.Body); err != nil {
			return errors.Wrapf(err, "copy %q to disk after %d bytes", m.URL(), n)
		}

		if err := resp.Body.Close(); err != nil {
			return errors.Wrap(err, "closing resp body")
		}
	}
	return nil
}
