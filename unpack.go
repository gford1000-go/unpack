package unpack

import (
	"encoding/json"
	"errors"
	"sort"
)

// Unpackable instances provide the ability to assign their name
// using the attribute name of the enclosing JSON object
type Unpackable[T any] interface {
	*T
	SetName(name string)
	GetName() string
}

// Unmarshal returns the slice of Unpackable instances within a JSON objects
// The Unpackable must be a pointer type implementation of the interface.
func Unmarshal[T any, PT Unpackable[T]](b []byte) ([]PT, error) {

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

	newT := func() PT {
		return new(T)
	}

	var ret = []PT{}

	for _, items := range m {

		// Sorting on the keys generates a deterministic return ordering
		sortedKeys := sort.StringSlice{}
		for k := range items {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Sort(sortedKeys)

		for _, name := range sortedKeys {
			item := items[name]
			b, err := json.Marshal(item) // Not ideal obvs ...
			if err != nil {
				return nil, err
			}

			// ... but easiest way to obtain the byte slice
			// to parse into actual structure
			r := newT()
			if err := json.Unmarshal(b, r); err != nil {
				return nil, err
			}
			r.SetName(name)

			ret = append(ret, r)
		}
	}

	return ret, nil
}

// Marshal encodes the slice of Unpackable instances to JSON
func Marshal[T any, PT Unpackable[T]](name string, data []PT) ([]byte, error) {

	m := map[string]PT{}

	for _, d := range data {
		m[d.GetName()] = d
	}

	if len(name) > 0 {
		mm := map[string]map[string]PT{}
		mm[name] = m
		return json.Marshal(mm)
	}

	return json.Marshal(m)
}
