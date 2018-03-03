package pkg

import (
	"archive/tar"
	"compress/bzip2"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/openpgp"
	yaml "gopkg.in/yaml.v2"

	"mcquay.me/fs"
	"mcquay.me/pm"

	"mcquay.me/pm/keyring"
)

var validNames map[string]bool
var crypto []string

func init() {
	validNames = map[string]bool{
		"root.tar.bz2":     true,
		"meta.yaml":        true,
		"bin/pre-install":  true,
		"bin/post-install": true,
		"bin/pre-upgrade":  true,
		"bin/post-upgrade": true,
		"bin/pre-remove":   true,
		"bin/post-remove":  true,
	}

	crypto = []string{
		"bom.sha256",
		"manifest.sha256",
		"manifest.sha256.asc",
	}
}

// Create traverses the contents of dir and emits a valid pkg, signed by id
func Create(key *openpgp.Entity, dir string) error {
	if !fs.Exists(dir) {
		return fmt.Errorf("%q: doesn't exist", dir)
	}
	if !fs.IsDir(dir) {
		return fmt.Errorf("%q: is not a directory", dir)
	}

	if err := clean(dir); err != nil {
		return errors.Wrap(err, "cleaning up directory")
	}

	files := []string{}
	found := map[string]bool{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if path == dir {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		p := strings.TrimPrefix(path, dir)[1:]

		if _, ok := validNames[p]; !ok {
			return fmt.Errorf("unxpected filename: %q", p)
		}

		if strings.HasPrefix(p, "bin") && info.Mode()&0100 == 0 {
			return fmt.Errorf("%q is not executable", path)
		}
		files = append(files, p)
		found[p] = true
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "checking directory contents")
	}

	if _, ok := found["root.tar.bz2"]; !ok {
		return fmt.Errorf("did not find root.tar.bz2")
	}
	if _, ok := found["meta.yaml"]; !ok {
		return fmt.Errorf("did not find meta.yaml")
	}

	mf, err := os.Open(filepath.Join(dir, "meta.yaml"))
	if err != nil {
		return errors.Wrap(err, "opening metadata file")
	}
	md := pm.Meta{}
	if err := yaml.NewDecoder(mf).Decode(&md); err != nil {
		return errors.Wrap(err, "decoding metadata file")
	}
	if _, err := md.Valid(); err != nil {
		return errors.Wrap(err, "invalid metadata")
	}

	f, err := os.Open(filepath.Join(dir, "root.tar.bz2"))
	if err != nil {
		return errors.Wrap(err, "opening overlay tarball")
	}

	bom, err := os.Create(filepath.Join(dir, "bom.sha256"))
	if err != nil {
		return errors.Wrap(err, "creating bom.sha256")
	}
	tr := tar.NewReader(bzip2.NewReader(f))
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			errors.Wrap(err, "traversing tarball")
		}
		if hdr.FileInfo().IsDir() {
			continue
		}
		s := sha256.New()
		if c, err := io.Copy(s, tr); err != nil {
			return errors.Wrapf(err, "copy after %d bytes", c)
		}
		fmt.Fprintf(bom, "%x\t%s\n", s.Sum(nil), hdr.Name)
	}
	if err := bom.Close(); err != nil {
		return errors.Wrap(err, "closing bom")
	}

	files = append(files, "bom.sha256")
	sort.Strings(files)

	manifest, err := os.Create(filepath.Join(dir, "manifest.sha256"))
	if err != nil {
		return errors.Wrap(err, "creating manifest.sha256")
	}
	for _, fn := range files {
		full := filepath.Join(dir, fn)
		f, err := os.Open(full)
		if err != nil {
			return errors.Wrap(err, "opening file for manifest checksumming")
		}
		s := sha256.New()
		if c, err := io.Copy(s, f); err != nil {
			return errors.Wrapf(err, "copy after %d bytes", c)
		}
		if err := f.Close(); err != nil {
			return errors.Wrapf(err, "closing %q", f.Name())
		}
		fmt.Fprintf(manifest, "%x\t%s\n", s.Sum(nil), fn)
	}
	if err := manifest.Close(); err != nil {
		return errors.Wrap(err, "closing manifest")
	}

	sig, err := os.Create(filepath.Join(dir, "manifest.sha256.asc"))
	if err != nil {
		return errors.Wrap(err, "creating sig file")
	}
	mfi, err := os.Open(filepath.Join(dir, "manifest.sha256"))
	if err != nil {
		return errors.Wrap(err, "opening manifest")
	}
	if err := keyring.Sign(key, mfi, sig); err != nil {
		return errors.Wrap(err, "signing")
	}

	files = append(files, "manifest.sha256")
	files = append(files, "manifest.sha256.asc")
	sort.Strings(files)

	tn, pn := filepath.Split(dir)
	tn = filepath.Join(tn, fmt.Sprintf("%v-%v.pkg", pn, md.Version))
	tf, err := os.Create(tn)
	if err != nil {
		return errors.Wrap(err, "opening final .pkg")
	}

	tw := tar.NewWriter(tf)
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if path == dir {
			return nil
		}
		p := strings.TrimPrefix(path, dir)[1:]
		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return errors.Wrap(err, "file info header")
		}
		hdr.Name = p
		if err := tw.WriteHeader(hdr); err != nil {
			return errors.Wrapf(err, "writing tar header for %v", p)
		}

		// only need to do real writing for non-directories
		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return errors.Wrap(err, "opening file for tar creation")
		}
		if c, err := io.Copy(tw, f); err != nil {
			log.Printf("%+v", err == io.EOF)
			return errors.Wrapf(err, "copy after %d bytes", c)
		}
		if err := f.Close(); err != nil {
			return errors.Wrap(err, "closing file during tar creation")
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "traversing the prepared directory")
	}
	if err := tw.Close(); err != nil {
		return errors.Wrap(err, "closing tar writer")
	}

	if err := tf.Close(); err != nil {
		return errors.Wrap(err, "closing final .pkg")
	}

	return nil
}

func clean(root string) error {
	for _, f := range crypto {
		path := filepath.Join(root, f)
		if !fs.Exists(path) {
			continue
		}
		if err := os.Remove(path); err != nil {
			return errors.Wrapf(err, "removing %q", path)
		}
	}
	return nil
}
