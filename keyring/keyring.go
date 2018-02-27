package keyring

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"

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

// Export prints pubkey information associated with email to w.
func Export(root string, w io.Writer, email string) error {
	if err := ensureDir(root); err != nil {
		return errors.Wrap(err, "can't find or create pgp dir")
	}
	srn, prn := getNames(root)
	_, pubs, err := getELs(srn, prn)
	if err != nil {
		return errors.Wrap(err, "getting existing keyrings")
	}

	e, err := findKey(pubs, email)
	if err != nil {
		return errors.Wrap(err, "find key")
	}

	aw, err := armor.Encode(w, openpgp.PublicKeyType, nil)
	if err != nil {
		return errors.Wrap(err, "creating armor encoder")
	}

	if err := e.Serialize(aw); err != nil {
		return errors.Wrap(err, "serializing key")
	}
	if err := aw.Close(); err != nil {
		return errors.Wrap(err, "closing armor encoder")
	}
	fmt.Fprintf(w, "\n")
	return nil
}

// Import parses public key information from w and adds it to the public
// keyring.
func Import(root string, w io.Reader) error {
	el, err := openpgp.ReadArmoredKeyRing(w)
	if err != nil {
		return errors.Wrap(err, "reading keyring")
	}

	if err := ensureDir(root); err != nil {
		return errors.Wrap(err, "can't find or create pgp dir")
	}
	srn, prn := getNames(root)
	_, pubs, err := getELs(srn, prn)
	if err != nil {
		return errors.Wrap(err, "getting existing keyrings")
	}

	foreign := openpgp.EntityList{}
	exist := map[uint64]bool{}
	for _, p := range pubs {
		exist[p.PrimaryKey.KeyId] = true
	}

	for _, e := range el {
		if _, ok := exist[e.PrimaryKey.KeyId]; !ok {
			foreign = append(foreign, e)
		}
	}
	if len(foreign) < 1 {
		return errors.New("no new key material found")
	}

	pubs = append(pubs, foreign...)

	pr, err := os.Create(prn)
	if err != nil {
		return errors.Wrap(err, "opening pubring")
	}
	for _, e := range pubs {
		if err := e.Serialize(pr); err != nil {
			return errors.Wrapf(err, "serializing %v", e.PrimaryKey.KeyIdString())
		}
	}
	if err := pr.Close(); err != nil {
		return errors.Wrap(err, "closing pubring")
	}
	return nil
}

// Sign takes an id and a reader and writes the signature for that id to sig.
func Sign(root, id string, in io.Reader, sig io.Writer) error {
	if err := ensureDir(root); err != nil {
		return errors.Wrap(err, "can't find or create pgp dir")
	}
	srn, prn := getNames(root)
	secs, _, err := getELs(srn, prn)
	if err != nil {
		return errors.Wrap(err, "getting existing keyrings")
	}
	e, err := findKey(secs, id)
	if err != nil {
		return errors.Wrapf(err, "finding key %q", id)
	}
	if err := openpgp.ArmoredDetachSign(sig, e, in, nil); err != nil {
		return errors.Wrap(err, "armored detach sign")
	}
	fmt.Fprintf(sig, "\n")
	return nil
}

// Verify verifies a file's deatched signature.
func Verify(root string, file, sig io.Reader) error {
	if err := ensureDir(root); err != nil {
		return errors.Wrap(err, "can't find or create pgp dir")
	}
	srn, prn := getNames(root)
	_, pubs, err := getELs(srn, prn)
	if err != nil {
		return errors.Wrap(err, "getting existing keyrings")
	}
	if _, err = openpgp.CheckArmoredDetachedSignature(pubs, file, sig); err != nil {
		return errors.Wrap(err, "check sig")
	}
	return nil
}

// Remove removes public key information for a given id.
//
// It skips public keys that have matching secret keys, and does not effect
// private keys.
func Remove(root string, id string) error {
	if err := ensureDir(root); err != nil {
		return errors.Wrap(err, "can't find or create pgp dir")
	}
	srn, prn := getNames(root)
	secs, pubs, err := getELs(srn, prn)
	if err != nil {
		return errors.Wrap(err, "getting existing keyrings")
	}
	victim, err := findKey(pubs, id)
	if err != nil {
		return errors.Wrapf(err, "finding key %q", id)
	}

	pr, err := os.Create(prn)
	if err != nil {
		return errors.Wrap(err, "opening pubring")
	}
	var rerr error
	for _, p := range pubs {
		if victim.PrimaryKey.KeyId == p.PrimaryKey.KeyId {
			if len(secs.KeysById(victim.PrimaryKey.KeyId)) == 0 {
				continue
			}
			rerr = fmt.Errorf("skipping pubkey with matching privkey: %v", p.PrimaryKey.KeyIdShortString())
		}

		if err := p.Serialize(pr); err != nil {
			return errors.Wrapf(err, "serializing %v", p.PrimaryKey.KeyIdString())
		}
	}
	if err := pr.Close(); err != nil {
		return errors.Wrap(err, "closing pubring")
	}

	return rerr
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

func findKey(el openpgp.EntityList, id string) (*openpgp.Entity, error) {
	var e *openpgp.Entity
	if strings.Contains(id, "@") {
		es := openpgp.EntityList{}
		for _, p := range el {
			for _, v := range p.Identities {
				if id == v.UserId.Email {
					es = append(es, p)
				}
			}
		}
		if len(es) == 1 {
			return es[0], nil
		}
		if len(es) > 1 {
			return nil, errors.New("too many keys matched; try searching by key id?")
		}
	} else {
		for _, p := range el {
			if id == p.PrimaryKey.KeyIdShortString() {
				return p, nil
			}
		}
	}
	return e, fmt.Errorf("key %q not found", id)
}
