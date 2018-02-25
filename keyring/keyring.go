package keyring

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/openpgp"

	"mcquay.me/fs"
)

// NewKeyPair creates and adds a new OpenPGP keypair to an existing keyring.
func NewKeyPair(root, name, email string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	if email == "" {
		return errors.New("email cannot be empty")
	}
	if strings.ContainsAny(email, "()<>\x00") {
		return fmt.Errorf("email %q contains invalid chars", email)
	}
	if err := ensureDir(root); err != nil {
		return errors.Wrap(err, "can't find or create pgp dir")
	}
	srn, prn := getNames(root)
	secs, pubs, err := getELs(srn, prn)
	if err != nil {
		return errors.Wrap(err, "getting existing keyrings")
	}

	fresh, err := openpgp.NewEntity(name, "pm", email, nil)
	if err != nil {
		errors.Wrap(err, "new entity")
	}

	pr, err := os.Create(prn)
	if err != nil {
		return errors.Wrap(err, "opening pubring")
	}
	sr, err := os.Create(srn)
	if err != nil {
		return errors.Wrap(err, "opening secring")
	}

	for _, e := range secs {
		if err := e.SerializePrivate(sr, nil); err != nil {
			return errors.Wrapf(err, "serializing old private key: %v", e.PrimaryKey.KeyIdString())
		}
	}
	// order is critical here; if we don't serialize the private key of fresh
	// first, the later steps fail.
	if err := fresh.SerializePrivate(sr, nil); err != nil {
		return errors.Wrapf(err, "serializing fresh private %v", fresh.PrimaryKey.KeyIdString())
	}
	if err := sr.Close(); err != nil {
		return errors.Wrap(err, "closing secring")
	}

	for _, e := range pubs {
		if err := e.Serialize(pr); err != nil {
			return errors.Wrapf(err, "serializing %v", e.PrimaryKey.KeyIdString())
		}
	}
	if err := fresh.Serialize(pr); err != nil {
		return errors.Wrapf(err, "serializing %v", fresh.PrimaryKey.KeyIdString())
	}
	if err := pr.Close(); err != nil {
		return errors.Wrap(err, "closing pubring")
	}
	return nil
}

// ListKeys prints keyring information to w.
func ListKeys(root string, w io.Writer) error {
	if err := ensureDir(root); err != nil {
		return errors.Wrap(err, "can't find or create pgp dir")
	}
	srn, prn := getNames(root)
	secs, pubs, err := getELs(srn, prn)
	if err != nil {
		return errors.Wrap(err, "getting existing keyrings")
	}
	for _, s := range secs {
		names := []string{}
		for _, v := range s.Identities {
			names = append(names, v.Name)
		}
		fmt.Fprintf(w, "sec: %+v:\t%v\n", s.PrimaryKey.KeyIdShortString(), strings.Join(names, ","))
	}
	for _, p := range pubs {
		names := []string{}
		for _, v := range p.Identities {
			names = append(names, v.Name)
		}
		fmt.Fprintf(w, "pub: %+v:\t%v\n", p.PrimaryKey.KeyIdShortString(), strings.Join(names, ","))
	}
	return nil
}

func pGPDir(root string) string {
	return filepath.Join(root, "var", "lib", "pm", "pgp")
}

func ensureDir(root string) error {
	d := pGPDir(root)
	if !fs.Exists(d) {
		if err := os.MkdirAll(d, 0700); err != nil {
			return errors.Wrap(err, "mk pgp dir")
		}
	}
	return nil
}

func getNames(root string) (string, string) {
	srn := filepath.Join(pGPDir(root), "secring.gpg")
	prn := filepath.Join(pGPDir(root), "pubring.gpg")
	return srn, prn
}

func getELs(secring, pubring string) (openpgp.EntityList, openpgp.EntityList, error) {
	var sr, pr openpgp.EntityList
	if fs.Exists(secring) {
		f, err := os.Open(secring)
		if err != nil {
			return nil, nil, errors.Wrap(err, "opening secring")
		}
		sr, err = openpgp.ReadKeyRing(f)
		if err != nil {
			return nil, nil, errors.Wrap(err, "read sec key ring")
		}
		if err := f.Close(); err != nil {
			return nil, nil, errors.Wrap(err, "closing keyring")
		}
	}

	if fs.Exists(pubring) {
		f, err := os.Open(pubring)
		if err != nil {
			return nil, nil, errors.Wrap(err, "opening pubring")
		}
		pr, err = openpgp.ReadKeyRing(f)
		if err != nil {
			return nil, nil, errors.Wrap(err, "read pub key ring")
		}
		if err := f.Close(); err != nil {
			return nil, nil, errors.Wrap(err, "closing keyring")
		}
	}
	return sr, pr, nil
}
