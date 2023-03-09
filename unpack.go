package unpack

import (
	"encoding/json"
	"errors"
)

// Unpackable instances provide the ability to assign their name
// using the attribute name of the enclosing JSON object
type Unpackable interface {
	SetName(name string)
}

// UnpackableFactory creates instances of Unpackable
type UnpackableFactory interface {
	New() Unpackable
}

// Unpack returns the slice of Unpackable instances within a JSON objects
// The Unpackable must be a pointer type implementation of the interface.
func Unpack[F UnpackableFactory](b []byte, fact F) ([]Unpackable, error) {

	/*
		The JSON structure should have been of the form:

		{
			<any attribute name - ignored> : {
				<unpackable name "X"> : { .... },
				<unpackable name "Y"> : { .... },
				...
				<unpackable name "Z"> : { .... }
			}
		}

		Each unpackable name is therefore unique.

		Each JSON object is expected to be the same structure

		Exit if the structure is not well formed
	*/
	var m map[string]map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}

	// Should only have a single entry in the outer map
	if len(m) != 1 {
		return nil, errors.New("incorrectly formed JSON")
	}

	var ret = []Unpackable{}

	for _, items := range m {
		for name, item := range items {
			b, err := json.Marshal(item) // Not ideal obvs ...
			if err != nil {
				return nil, err
			}

			// ... but easiest way to obtain the byte slice
			// to parse into actual structure
			r := fact.New()
			if err := json.Unmarshal(b, r); err != nil {
				return nil, err
			}
			r.SetName(name)

			ret = append(ret, r)
		}
	}

	return ret, nil
}
