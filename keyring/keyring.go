package keyring

import (
	"log"

	"github.com/pkg/errors"
	"golang.org/x/crypto/openpgp"
)

func NewKeyPair(root, name, email string) error {
	e, err := openpgp.NewEntity(name, "pm", email, nil)
	if err != nil {
		errors.Wrap(err, "new entity")
	}
	log.Printf("%+v", e)
	return errors.New("NYI")
}
