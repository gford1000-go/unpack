package unpack

type structType int

func (j structType) isValid() bool {
	return j > unknownStructType && j < invalidJSONStructType
}

const (
	unknownStructType structType = iota
	namedItemMap
	anonymousItemMap
	structuredMap
	invalidJSONStructType
)

// Ordering specifies how data items should be sorted
type Ordering int

func (o Ordering) isValid() bool {
	return o > UnknownOrdering && o < InvalidOrdering
}

const (
	UnknownOrdering Ordering = iota
	Ascending
	Descending
	InvalidOrdering
)

// Options allow the behaviour of marshaling and unmarshaling to be modified
type Options[T any, PT Unpackable[T]] struct {
	structType structType
	// NewFn provides the ability to initialise instances of T prior to unmarshaling JSON
	NewFn func() PT
	// Ordering defines how the data items will be sorted, using their names
	Ordering Ordering
}

// withStructType allows the type of struct to be specified
// Default: namedItemMap
func withStructType[T any, PT Unpackable[T]](j structType) func(*Options[T, PT]) {
	return func(o *Options[T, PT]) {
		if j.isValid() {
			o.structType = j
		}
	}
}

// WithNewFn allows instance creation to be overridden
func WithNewFn[T any, PT Unpackable[T]](fn func() PT) func(*Options[T, PT]) {
	return func(o *Options[T, PT]) {
		if fn != nil {
			o.NewFn = fn
		}
	}
}

// WithOrdering specifies how data items should be sorted when being unmashaled
// Default: Ascending
func WithOrdering[T any, PT Unpackable[T]](ordering Ordering) func(*Options[T, PT]) {
	return func(o *Options[T, PT]) {
		if ordering.isValid() {
			o.Ordering = ordering
		}
	}
}
