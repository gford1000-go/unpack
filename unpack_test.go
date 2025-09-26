package unpack

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type myTestType struct {
	n string
}

func (t *myTestType) SetName(name string) {
	t.n = name
}

func (t *myTestType) GetName() string {
	return t.n
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

		u, err := UnmarshalWithName[myTestType]("a", []byte(test.json))
		if err != nil {
			if test.parseable {
				t.Fatalf("Unexpected parse failure for test %d: %v", i, err)
			} else {
				continue
			}
		}

		assert.Equal(t, len(test.names), len(u))

		for i, uu := range u {
			assert.Equal(t, test.names[i], uu.n)
		}
	}
}

type myCountryDetails struct {
	Name       string
	Capital    string         `json:"capital"`
	Population map[string]int `json:"population"`
}

func (c *myCountryDetails) SetName(name string) {
	c.Name = name
}

func (c *myCountryDetails) GetName() string {
	return c.Name
}

func ExampleUnmarshal() {

	b := []byte(`{"UK":{"capital":"London","population":{"London":12000000}},"US":{"capital":"Washington DC","population":{"Washington DC":9500000}}}`)

	countries, _ := Unmarshal[myCountryDetails](b)
	for _, country := range countries {
		fmt.Println(country.Name, country.Capital, country.Population[country.Capital])
	}

	// Output:
	// UK London 12000000
	// US Washington DC 9500000
}

func ExampleMarshalWithName() {

	countries := []*myCountryDetails{
		{
			Name:    "UK",
			Capital: "London",
			Population: map[string]int{
				"London": 10000000,
			},
		},
		{
			Name:    "US",
			Capital: "Washington DC",
			Population: map[string]int{
				"Washington DC": 95000000,
			},
		},
	}

	b, _ := MarshalWithName("countries", countries)
	fmt.Println(string(b))

	// Output:
	// {"countries":{"UK":{"Name":"UK","capital":"London","population":{"London":10000000}},"US":{"Name":"US","capital":"Washington DC","population":{"Washington DC":95000000}}}}
}

func ExampleMarshal() {

	countries := []*myCountryDetails{
		{
			Name:    "UK",
			Capital: "London",
			Population: map[string]int{
				"London": 10000000,
			},
		},
		{
			Name:    "US",
			Capital: "Washington DC",
			Population: map[string]int{
				"Washington DC": 95000000,
			},
		},
	}

	b, _ := Marshal(countries)
	fmt.Println(string(b))

	// Output:
	// {"UK":{"Name":"UK","capital":"London","population":{"London":10000000}},"US":{"Name":"US","capital":"Washington DC","population":{"Washington DC":95000000}}}
}

func TestMarshal(t *testing.T) {

	data := []byte(`{"UK":{"Name":"UK","capital":"London","population":{"London":10000000}},"US":{"Name":"US","capital":"Washington DC","population":{"Washington DC":95000000}}}`)

	objs, err := Unmarshal[myCountryDetails](data)
	if err != nil {
		t.Fatalf("unexpected Unmarshal error: %v", err)
	}

	data1, err := Marshal(objs)
	if err != nil {
		t.Fatalf("unexpected Marshal error: %v", err)
	}

	if !bytes.Equal(data, data1) {
		t.Fatalf("mismatch: expected: %s, got: %s", string(data), string(data1))
	}
}

func TestMarshalWithName(t *testing.T) {

	name := "countries"
	data := []byte(fmt.Sprintf(`{"%s":{"UK":{"Name":"UK","capital":"London","population":{"London":10000000}},"US":{"Name":"US","capital":"Washington DC","population":{"Washington DC":95000000}}}}`, name))

	objs, err := UnmarshalWithName[myCountryDetails](name, data)
	if err != nil {
		t.Fatalf("unexpected Unmarshal error: %v", err)
	}

	data1, err := MarshalWithName(name, objs)
	if err != nil {
		t.Fatalf("unexpected Marshal error: %v", err)
	}

	if !bytes.Equal(data, data1) {
		t.Fatalf("mismatch: expected: %s, got: %s", string(data), string(data1))
	}
}
