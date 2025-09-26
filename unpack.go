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

	var m map[string]any = nil
	var mm map[string]map[string]any = nil

	defer func() {
		switch o.structType {
		case anonymousItemMap:
			if m != nil {
				releaseMap(m)
			}
		case namedItemMap:
			if mm != nil {
				releaseMap2Map(mm)
			}
		}
	}()

	switch o.structType {
	case namedItemMap:
		mm = acquireMap2Map()
		if err := json.Unmarshal(b, &mm); err != nil {
			return nil, err
		}
		if nm, ok := mm[name]; !ok {

		} else {
			m = nm
		}
	case anonymousItemMap:
		m = acquireMap()
		if err := json.Unmarshal(b, &m); err != nil {
			return nil, err
		}
	}

	// Sorting on the keys generates a deterministic return ordering
	sortedKeys := make(sort.StringSlice, 0, len(m))
	for k := range m {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Sort(sortedKeys)

	var ret = make([]PT, 0, len(m))

	for _, name := range sortedKeys {

		r := o.NewFn()
		if err := mapToStruct(m[name].(map[string]any), r); err != nil {

			// mapToStruct could have edge case failures, in which case
			// use json roundtrip to try to decode
			r = o.NewFn()
			if err := roundTripToStruct(m[name], r); err != nil {
				return nil, err
			}
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

	m := map[string]any{}

	for _, d := range data {
		m[d.GetName()] = d
	}

	switch o.structType {
	case namedItemMap:
		mm := acquireMap2Map()
		defer releaseMap2Map(mm)
		mm[name] = m
		return json.Marshal(mm)
	case anonymousItemMap:
		return json.Marshal(m)
	default:
		panic(fmt.Sprintf("unsupported value of Options.structType provided (%d)", o.structType))
	}
}
