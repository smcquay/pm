package pm

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

func TestValid(t *testing.T) {
	tests := []struct {
		label string
		m     Meta
		ok    bool
		err   error
	}{
		{
			label: "valid",
			m: Meta{
				Name:        "heat",
				Version:     "1.1.0",
				Description: "some description",
			},
			ok: true,
		},
		{
			label: "missing name",
			m: Meta{
				Version:     "1.1.0",
				Description: "some description",
			},
			err: errors.New("name"),
		},
		{
			label: "missing version",
			m: Meta{
				Name:        "heat",
				Description: "some description",
			},
			err: errors.New("version"),
		},
		{
			label: "missing description",
			m: Meta{
				Name:    "heat",
				Version: "1.1.0",
			},
			err: errors.New("description"),
		},
	}

	for _, test := range tests {
		t.Run(test.label, func(t *testing.T) {
			ok, err := test.m.Valid()
			if got, want := ok, test.ok; got != want {
				t.Fatalf("validity: got %v, want %v", got, want)
			}

			if got, want := err, test.err; (err == nil) != (test.err == nil) {
				t.Fatalf("error: got %v, want %v", got, want)
			}
		})
	}
}

func TestJsonRoundTrip(t *testing.T) {
	b := Meta{
		Name:        "heat",
		Version:     "1.1.0",
		Description: "make heat using cpus",
	}

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(&b); err != nil {
		t.Fatalf("encode: %v", err)
	}

	a := Meta{}
	if err := json.NewDecoder(buf).Decode(&a); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if b != a {
		t.Fatalf("a != b: %v != %v", a, b)
	}
}
