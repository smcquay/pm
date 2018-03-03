package pm

import "errors"

type Meta struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Namespace   string `json:"namespace"`
}

func (m *Meta) Valid() (bool, error) {
	if m.Name == "" {
		return false, errors.New("name cannot be empty")
	}
	if m.Version == "" {
		return false, errors.New("version cannot be empty")
	}
	if m.Description == "" {
		return false, errors.New("description cannot be empty")
	}
	return true, nil
}
