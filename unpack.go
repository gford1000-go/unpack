package unpack

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
)

// Unpackable instances provide the ability to assign their name
// using the attribute name of the enclosing JSON object
type Unpackable[T any] interface {
	*T
	SetName(name string)
	GetName() string
}

// ErrNoNameSpecified is returned if a ...WithName function is called with an empty string name
var ErrNoNameSpecified = errors.New("name must be provided")

// UnmarshalWithName will decode the inner named objects into instances of T,
// from the object specified by the name in the outer map
func UnmarshalWithName[T any, PT Unpackable[T]](name string, b []byte, opts ...func(*Options[T, PT])) ([]PT, error) {
	/*
		The JSON structure is a map of maps, of the form:

		{
			<element with specified name> : {
				<unpackable name "X"> : { .... },
				<unpackable name "Y"> : { .... },
				...
				<unpackable name "Z"> : { .... }
			}
		}

		Each unpackable name is therefore unique.

		Each JSON object is expected to be the same structure

		Exit if the structure is not well formed, or name not provided
	*/

	if len(name) == 0 {
		return nil, ErrNoNameSpecified
	}
	return unmarshal(name, b, append(opts, withStructType[T, PT](namedItemMap))...)
}

// UnmarshalWithName will decode the map into named objects into instances of T
func Unmarshal[T any, PT Unpackable[T]](b []byte, opts ...func(*Options[T, PT])) ([]PT, error) {
	/*
		The JSON structure is a simple map, i.e. of the form:

		{
			<unpackable name "X"> : { .... },
			<unpackable name "Y"> : { .... },
			...
			<unpackable name "Z"> : { .... }
		}

		Each unpackable name is therefore unique.

		Each JSON object is expected to be the same structure

		Exit if the structure is not well formed
	*/

	return unmarshal("", b, append(opts, withStructType[T, PT](anonymousItemMap))...)
}

func newT[T any, PT Unpackable[T]]() PT {
	return new(T)
}

// Unmarshal returns the slice of Unpackable instances within a JSON objects
// The Unpackable must be a pointer type implementation of the interface.
func unmarshal[T any, PT Unpackable[T]](name string, b []byte, opts ...func(*Options[T, PT])) ([]PT, error) {

	o := Options[T, PT]{
		structType: namedItemMap,
		NewFn:      newT[T, PT],
	}
	for _, opt := range opts {
		opt(&o)
	}

	var m = map[string]any{}

	switch o.structType {
	case namedItemMap:
		mm := map[string]map[string]any{}
		if err := json.Unmarshal(b, &mm); err != nil {
			return nil, err
		}
		if nm, ok := mm[name]; !ok {

		} else {
			m = nm
		}
	case anonymousItemMap:
		if err := json.Unmarshal(b, &m); err != nil {
			return nil, err
		}
	}

	// Sorting on the keys generates a deterministic return ordering
	sortedKeys := sort.StringSlice{}
	for k := range m {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Sort(sortedKeys)

	var ret = []PT{}

	for _, name := range sortedKeys {
		item := m[name]
		b, err := json.Marshal(item) // Not ideal obvs ...
		if err != nil {
			return nil, err
		}

		// ... but easiest way to obtain the byte slice
		// to parse into actual structure
		r := o.NewFn()
		if err := json.Unmarshal(b, r); err != nil {
			return nil, err
		}
		r.SetName(name)

		ret = append(ret, r)
	}

	return ret, nil
}

// Marshal encodes the slice of Unpackable instances to a JSON anonymous map
func Marshal[T any, PT Unpackable[T]](data []PT, opts ...func(*Options[T, PT])) ([]byte, error) {
	return marshal("", data, append(opts, withStructType[T, PT](anonymousItemMap))...)
}

// MarshalWithName encodes the slice of Unpackable instances to a JSON named map
func MarshalWithName[T any, PT Unpackable[T]](name string, data []PT, opts ...func(*Options[T, PT])) ([]byte, error) {
	if len(name) == 0 {
		return nil, ErrNoNameSpecified
	}
	return marshal(name, data, append(opts, withStructType[T, PT](namedItemMap))...)
}

func marshal[T any, PT Unpackable[T]](name string, data []PT, opts ...func(*Options[T, PT])) ([]byte, error) {

	o := Options[T, PT]{
		structType: namedItemMap,
	}
	for _, opt := range opts {
		opt(&o)
	}

	m := map[string]PT{}

	for _, d := range data {
		m[d.GetName()] = d
	}

	switch o.structType {
	case namedItemMap:
		mm := map[string]map[string]PT{}
		mm[name] = m
		return json.Marshal(mm)
	case anonymousItemMap:
		return json.Marshal(m)
	default:
		panic(fmt.Sprintf("unsupported value of Options.structType provided (%d)", o.structType))
	}
}
