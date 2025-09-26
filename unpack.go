package unpack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
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
func UnmarshalWithName[T any, PT Unpackable[T]](ctx context.Context, name string, b []byte, opts ...func(*Options[T, PT])) ([]PT, error) {
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
	sm, err := unmarshal[any, T, PT](ctx, "", name, b, append(opts, withStructType[T, PT](namedItemMap))...)
	if err != nil {
		return nil, err
	}
	return sm.Data, nil
}

// UnmarshalWithName will decode the map into named objects into instances of T
func Unmarshal[T any, PT Unpackable[T]](ctx context.Context, b []byte, opts ...func(*Options[T, PT])) ([]PT, error) {
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

	sd, err := unmarshal[any, T, PT](ctx, "", "", b, append(opts, withStructType[T, PT](anonymousItemMap))...)
	if err != nil {
		return nil, err
	}
	return sd.Data, nil
}

func newT[T any, PT Unpackable[T]]() PT {
	return new(T)
}

func newM[M any]() *M {
	return new(M)
}

// ErrMetaNameNotFound is returned if the specified meta data name is not in the supplied []byte slice
var ErrMetaNameNotFound = errors.New("meta name is not found")

// ErrDataNameNotFound is returned if the specified data name is not in the supplied []byte slice
var ErrDataNameNotFound = errors.New("data name is not found")

// Unmarshal returns the slice of Unpackable instances within a JSON objects
// The Unpackable must be a pointer type implementation of the interface.
func unmarshal[M, T any, PT Unpackable[T]](ctx context.Context, metaName, dataName string, b []byte, opts ...func(*Options[T, PT])) (*StructuredData[M, T, PT], error) {

	o := Options[T, PT]{
		structType: namedItemMap,
		NewFn:      newT[T, PT],
		Ordering:   Ascending,
	}
	for _, opt := range opts {
		opt(&o)
	}

	var mMeta map[string]any = nil
	var mData map[string]any = nil
	var mm map[string]map[string]any = nil

	defer func() {
		switch o.structType {
		case anonymousItemMap:
			if mData != nil {
				releaseMap(mData)
			}
		case namedItemMap, structuredMap:
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
		if nm, ok := mm[dataName]; !ok {
			return nil, ErrDataNameNotFound
		} else {
			mData = nm
		}
	case structuredMap:
		mm = acquireMap2Map()
		if err := json.Unmarshal(b, &mm); err != nil {
			return nil, err
		}
		if nm, ok := mm[metaName]; !ok {
			return nil, ErrMetaNameNotFound
		} else {
			mMeta = nm
		}
		if nm, ok := mm[dataName]; !ok {
			return nil, ErrDataNameNotFound
		} else {
			mData = nm
		}
	case anonymousItemMap:
		mData = acquireMap()
		if err := json.Unmarshal(b, &mData); err != nil {
			return nil, err
		}
	}

	var meta *M = nil
	if mMeta != nil {
		meta = newM[M]()
		if err := mapToStruct(mMeta, meta); err != nil {
			return nil, err
		}
	}

	// Sorting on the keys generates a deterministic return ordering
	sortedKeys := make(sort.StringSlice, 0, len(mData))
	for k := range mData {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Sort(sortedKeys)
	if o.Ordering == Descending {
		slices.Reverse(sortedKeys)
	}

	var ptData = make([]PT, 0, len(mData))

	for _, name := range sortedKeys {

		r := o.NewFn()
		if err := mapToStruct(mData[name].(map[string]any), r); err != nil {

			// mapToStruct could have edge case failures, in which case
			// use json roundtrip to try to decode
			r = o.NewFn()
			if err := roundTripToStruct(mData[name], r); err != nil {
				return nil, err
			}
		}

		r.SetName(name)

		ptData = append(ptData, r)
	}

	return &StructuredData[M, T, PT]{
		Meta: meta,
		Data: ptData,
	}, nil
}

// Marshal encodes the slice of Unpackable instances to a JSON anonymous map
func Marshal[T any, PT Unpackable[T]](ctx context.Context, data []PT, opts ...func(*Options[T, PT])) ([]byte, error) {
	return marshal(ctx, "", data, append(opts, withStructType[T, PT](anonymousItemMap))...)
}

// MarshalWithName encodes the slice of Unpackable instances to a JSON named map
func MarshalWithName[T any, PT Unpackable[T]](ctx context.Context, name string, data []PT, opts ...func(*Options[T, PT])) ([]byte, error) {
	if len(name) == 0 {
		return nil, ErrNoNameSpecified
	}
	return marshal(ctx, name, data, append(opts, withStructType[T, PT](namedItemMap))...)
}

func marshal[T any, PT Unpackable[T]](ctx context.Context, name string, data []PT, opts ...func(*Options[T, PT])) ([]byte, error) {

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

// StructuredData is used to decode JSON where there are two elements in an outer map, one
// of which is metadata and the other contains a map of actual data
type StructuredData[M, T any, PT Unpackable[T]] struct {
	Meta *M
	Data []PT
}

// UnmarshalStructuredData decodes a JSON map of maps into a StructuredData instance, using the
// names to identify the metadata and data objects
func UnmarshalStructuredData[M, T any, PT Unpackable[T]](ctx context.Context, metaName, dataName string, b []byte, opts ...func(*Options[T, PT])) (*StructuredData[M, T, PT], error) {
	return unmarshal[M](ctx, metaName, dataName, b, append(opts, withStructType[T, PT](structuredMap))...)
}
