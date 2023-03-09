package unpack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type tt struct {
	n string
}

func (t *tt) SetName(name string) {
	t.n = name
}

type ttf struct{}

func (f ttf) New() Unpackable {
	return new(tt)
}

func TestUnpackAll(t *testing.T) {

	type tc struct {
		json      string
		parseable bool
		names     []string
	}

	tests := []tc{
		{
			json: `
{
	"a": {}
}
			`,
			parseable: true,
			names:     []string{},
		},
		{
			json: `
{
	"a": {
		"x": {}
	}
}
			`,
			parseable: true,
			names:     []string{"x"},
		},
		{
			json: `
{
	"a": {
		"x": {},
		"y": {}
	}
}
			`,
			parseable: true,
			names:     []string{"x", "y"},
		},
		{
			json: `
{
	"a": {
		"x": { "n": 1, "y":"hhsa"},
		"y": {}
	}
}
			`,
			parseable: true,
			names:     []string{"x", "y"},
		},
		{
			json: `
[
	{ "n": 1, "y":"hhsa"},
	{}
]
			`,
			parseable: false,
			names:     []string{},
		},
	}

	for i, test := range tests {

		u, err := Unpack([]byte(test.json), ttf{})
		if err != nil {
			if test.parseable {
				t.Fatalf("Unexpected parse failure for test %d: %v", i, err)
			} else {
				continue
			}
		}

		assert.Equal(t, len(test.names), len(u))

		for i, uu := range u {
			v := uu.(*tt)
			assert.Equal(t, test.names[i], v.n)
		}
	}
}
