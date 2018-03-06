package pkg

import (
	"archive/tar"
	"bufio"
	"compress/bzip2"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"mcquay.me/fs"
	"mcquay.me/pm"
	"mcquay.me/pm/db"
	"mcquay.me/pm/keyring"
)

const cache = "var/cache/pm"
const installed = "var/lib/pm/installed"

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
	installedDir := filepath.Join(root, installed)
	if !fs.Exists(installedDir) {
		if err := os.MkdirAll(installedDir, 0755); err != nil {
			return errors.Wrap(err, "creating non-existent cache dir")
		}
	}
	if !fs.IsDir(cacheDir) {
		return errors.Errorf("%q is not a directory!", cacheDir)
	}

	if err := download(cacheDir, ms); err != nil {
		return errors.Wrap(err, "downloading")
	}

	for _, m := range ms {
		if err := verifyManifestIntegrity(root, m); err != nil {
			return errors.Wrap(err, "verifying pkg integrity")
		}
		if err := expandPkgContents(root, m); err != nil {
			if err := os.RemoveAll(filepath.Join(root, installed, string(m.Name))); err != nil {
				err = errors.Wrap(err, "cleaning up")
			}
			return errors.Wrap(err, "verifying pkg contents")
		}

		if err := script(root, m, "pre-install"); err != nil {
			return errors.Wrap(err, "pre-install")
		}

		if err := expandRoot(root, m); err != nil {
			return errors.Wrap(err, "root expansion")
		}

		if err := script(root, m, "post-install"); err != nil {
			return errors.Wrap(err, "pre-install")
		}

		if err := os.Remove(filepath.Join(cacheDir, m.Pkg())); err != nil {
			return errors.Wrapf(err, "cleaning up pkg %v", m.Pkg())
		}
	}
	return nil
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

func verifyManifestIntegrity(root string, m pm.Meta) error {
	pn := filepath.Join(root, cache, m.Pkg())
	man, err := getReadCloser(pn, "manifest.sha256")
	if err != nil {
		return errors.Wrap(err, "getting manifest reader")
	}
	sig, err := getReadCloser(pn, "manifest.sha256.asc")
	if err != nil {
		return errors.Wrap(err, "getting manifest reader")
	}

	if err := keyring.Verify(root, man, sig); err != nil {
		return errors.Wrap(err, "verifying manifest")
	}
	if err := man.Close(); err != nil {
		return errors.Wrap(err, "closing manifest reader")
	}
	if err := sig.Close(); err != nil {
		return errors.Wrap(err, "closing manifest signature reader")
	}
	return nil
}

func expandPkgContents(root string, m pm.Meta) error {
	pn := filepath.Join(root, cache, m.Pkg())
	man, err := getReadCloser(pn, "manifest.sha256")
	if err != nil {
		return errors.Wrap(err, "getting manifest reader")
	}

	ip := filepath.Join(root, installed, string(m.Name))
	if err := os.MkdirAll(ip, 0755); err != nil {
		return errors.Wrapf(err, "making install dir for %q", m.Name)
	}

	cs := map[string]string{}
	s := bufio.NewScanner(man)
	for s.Scan() {
		elems := strings.Split(s.Text(), "\t")
		if len(elems) != 2 {
			return errors.Errorf("manifest format error; got %d elements, want 2", len(elems))
		}
		cs[elems[1]] = elems[0]
	}
	if err := man.Close(); err != nil {
		return errors.Wrap(err, "closing manifest reader")
	}
	if err := s.Err(); err != nil {
		return errors.Wrap(err, "scanning manifest")
	}

	pf, err := os.Open(pn)
	if err != nil {
		return errors.Wrap(err, "opening pkg file")
	}
	tr := tar.NewReader(pf)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.Wrap(err, "tar traversal")
		}

		if hdr.Name == "manifest.sha256" || hdr.Name == "manifest.sha256.asc" {
			continue
		}

		if hdr.FileInfo().IsDir() {
			if hdr.Name != "bin" {
				return errors.Errorf("%v is unexpected", hdr.Name)
			}
			if err := os.MkdirAll(filepath.Join(ip, hdr.Name), hdr.FileInfo().Mode()); err != nil {
				return errors.Wrapf(err, "mkdir for %v", hdr.Name)
			}
			continue
		}

		sha, ok := cs[hdr.Name]
		if !ok {
			return errors.Errorf("extra file %q found in tarfile!", hdr.Name)
		}

		name := filepath.Join(ip, hdr.Name)
		sr := sha256.New()
		var o io.WriteCloser
		o = close{ioutil.Discard}
		if hdr.Name != "root.tar.bz2" {
			f, err := os.OpenFile(filepath.Join(ip, hdr.Name), os.O_WRONLY|os.O_CREATE, hdr.FileInfo().Mode())
			if err != nil {
				return errors.Wrap(err, "open file in install dir")
			}
			o = f
		}

		w := io.MultiWriter(o, sr)

		if n, err := io.Copy(w, tr); err != nil {
			return errors.Wrapf(err, "copying file %q after %v bytes", hdr.Name, n)
		}

		if sha != fmt.Sprintf("%x", sr.Sum(nil)) {
			return errors.Errorf("%q checksum was incorrect", hdr.Name)
		}

		if err := o.Close(); err != nil {
			return errors.Wrapf(err, "closing %v", name)
		}
	}
	return nil
}

type tarSlurper struct {
	f  *os.File
	tr *tar.Reader
}

func (ts *tarSlurper) Close() error {
	return ts.f.Close()
}

func (ts *tarSlurper) Read(p []byte) (int, error) {
	return ts.tr.Read(p)
}

func getReadCloser(tn, fn string) (io.ReadCloser, error) {
	pf, err := os.Open(tn)
	if err != nil {
		return nil, errors.Wrap(err, "opening pkg file")
	}
	tr := tar.NewReader(pf)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errors.Wrap(err, "tar traversal")
		}
		if hdr.Name == fn {
			return &tarSlurper{pf, tr}, nil
		}
	}
	return nil, errors.Errorf("%q not found", fn)
}

// close should be used to wrap ioutil.Discard to give it a noop Close method.
type close struct {
	io.Writer
}

func (close) Close() error {
	return nil
}

func script(root string, m pm.Meta, name string) error {
	bin := filepath.Join(root, installed, string(m.Name), "bin", name)
	if !fs.Exists(bin) {
		return nil
	}
	cmd := exec.Command(bin)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func expandRoot(root string, m pm.Meta) error {

	bomn := filepath.Join(root, installed, string(m.Name), "bom.sha256")
	bf, err := os.Open(bomn)
	if err != nil {
		return errors.Wrap(err, "opening bom")
	}

	cs := map[string]string{}
	s := bufio.NewScanner(bf)
	for s.Scan() {
		elems := strings.Split(s.Text(), "\t")
		if len(elems) != 2 {
			return errors.Errorf("manifest format error; got %d elements, want 2", len(elems))
		}
		cs[elems[1]] = elems[0]
	}

	pn := filepath.Join(root, cache, m.Pkg())
	tbz, err := getReadCloser(pn, "root.tar.bz2")
	if err != nil {
		return errors.Wrap(err, "getting root.tar.bz2 reader")
	}
	tr := tar.NewReader(bzip2.NewReader(tbz))
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.Wrap(err, "tar traversal")
		}
		if hdr.FileInfo().IsDir() {
			d := filepath.Join(root, hdr.Name)
			if err := os.MkdirAll(d, hdr.FileInfo().Mode()); err != nil {
				return errors.Wrapf(err, "making directory %q", d)
			}
			continue
		}
		f, err := os.OpenFile(filepath.Join(root, hdr.Name), os.O_WRONLY|os.O_CREATE, hdr.FileInfo().Mode())
		if err != nil {
			return errors.Wrapf(err, "open output file %q", hdr.Name)
		}
		if n, err := io.Copy(f, tr); err != nil {
			return errors.Wrapf(err, "copy file %q after %v bytes", hdr.Name, n)
		}
		if err := f.Close(); err != nil {
			return errors.Wrapf(err, "closing %q", hdr.Name)
		}
	}
	return nil
}
