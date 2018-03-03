package pm

import (
	"errors"
	"testing"
)

func TestAvailableAdd(t *testing.T) {

	tests := []struct {
		label string
		m     Meta
		count int
		err   error
	}{
		{
			label: "good",
			m:     Meta{Name: "a", Version: "v1.0.0", Description: "test"},
			count: 1,
		},
		{
			label: "bad meta",
			m:     Meta{Name: "a"},
			count: 1,
			err:   errors.New("missing"),
		},
		{
			label: "dupe is last in",
			m:     Meta{Name: "a", Version: "v1.0.0", Description: "better version"},
			count: 1,
		},
		{
			label: "another good",
			m:     Meta{Name: "a", Version: "v1.0.0", Description: "better version"},
			count: 1,
		},
	}

	a := Available{}
	for _, test := range tests {
		t.Run(test.label, func(t *testing.T) {
			if err := a.Add(test.m); (err == nil) != (test.err == nil) {
				t.Fatalf("adding meta%v", err)
			}

			if got, want := len(a), test.count; got != want {
				t.Fatalf("unexpected length after Add: got %v, want %v", got, want)
			}
		})
	}

	if got, want := a["a"]["v1.0.0"].Description, "better version"; got != want {
		t.Fatalf("version: got %v, want %v", got, want)
	}
}

func TestAvailableUpdate(t *testing.T) {
	a := Available{}
	if err := a.Add(Meta{Name: "a", Version: "v1.0.0", Description: "test"}); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := a.Add(Meta{Name: "b", Version: "v2.0.0", Description: "test"}); err != nil {
		t.Fatalf("add: %v", err)
	}

	b := Available{}
	a.Update(b)
	if got, want := len(a), 2; got != want {
		t.Fatalf("len after empty update: got %v, want %v", got, want)
	}

	if err := b.Add(Meta{Name: "a", Version: "v1.0.0", Description: "test last in"}); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := b.Add(Meta{Name: "b", Version: "v2.1.0", Description: "test"}); err != nil {
		t.Fatalf("add: %v", err)
	}

	if err := a.Update(b); err != nil {
		t.Fatalf("update: %v", err)
	}
	if got, want := len(a), 2; got != want {
		t.Fatalf("len after update: got %v, want %v", got, want)
	}
	if got, want := len(a["a"]), 1; got != want {
		t.Fatalf("len after update: got %v, want %v", got, want)
	}
	if got, want := len(a["b"]), 2; got != want {
		t.Fatalf("len after update: got %v, want %v", got, want)
	}

	if got, want := a["a"]["v1.0.0"].Description, "test last in"; got != want {
		t.Fatalf("last in didn't override")
	}
}
