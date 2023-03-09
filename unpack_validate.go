package unpack

import valid "github.com/asaskevich/govalidator"

// UnpackAndValidate returns a validated set of Unpackables, where the
// validation to be performed is defined in the tag of each attribute
// see: https://pkg.go.dev/github.com/asaskevich/govalidator?utm_source=godoc
func UnpackAndValidate(b []byte, fact UnpackableFactory) ([]Unpackable, error) {

	unpackables, err := Unpack(b, fact)
	if err != nil {
		return nil, err
	}

	// Validate
	for _, unpackable := range unpackables {
		_, err = valid.ValidateStruct(unpackable)
		if err != nil {
			return nil, err
		}
	}

	return unpackables, nil
}
