// JSON saving and loading.

package aio

import "encoding/json"

// ToJSON saves a value in a JSON file.
func ToJSON(file string, a interface{}) error {
	f, err := Create(file)
	if err != nil {
		return err
	}
	e := json.NewEncoder(f)
	e.SetIndent("", "  ")
	if err := e.Encode(a); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// FromJSON loads a value from a JSON file.
func FromJSON(file string, a interface{}) error {
	f, err := Open(file)
	if err != nil {
		return err
	}
	e := json.NewDecoder(f)
	if err := e.Decode(a); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
