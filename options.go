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

// Options allow the behaviour of marshaling and unmarshaling to be modified
type Options[T any, PT Unpackable[T]] struct {
	structType structType
	// NewFn provides the ability to initialise instances of T prior to unmarshaling JSON
	NewFn func() PT
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
