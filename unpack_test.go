package unpack

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
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

		u, err := UnmarshalWithName[myTestType](context.Background(), "a", []byte(test.json))
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

	countries, _ := Unmarshal[myCountryDetails](context.Background(), b)
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

	b, _ := MarshalWithName(context.Background(), "countries", countries)
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

	b, _ := Marshal(context.Background(), countries)
	fmt.Println(string(b))

	// Output:
	// {"UK":{"Name":"UK","capital":"London","population":{"London":10000000}},"US":{"Name":"US","capital":"Washington DC","population":{"Washington DC":95000000}}}
}

func TestMarshal(t *testing.T) {

	data := []byte(`{"UK":{"Name":"UK","capital":"London","population":{"London":10000000}},"US":{"Name":"US","capital":"Washington DC","population":{"Washington DC":95000000}}}`)

	objs, err := Unmarshal[myCountryDetails](context.Background(), data)
	if err != nil {
		t.Fatalf("unexpected Unmarshal error: %v", err)
	}

	data1, err := Marshal(context.Background(), objs)
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

	objs, err := UnmarshalWithName[myCountryDetails](context.Background(), name, data)
	if err != nil {
		t.Fatalf("unexpected Unmarshal error: %v", err)
	}

	data1, err := MarshalWithName(context.Background(), name, objs)
	if err != nil {
		t.Fatalf("unexpected Marshal error: %v", err)
	}

	if !bytes.Equal(data, data1) {
		t.Fatalf("mismatch: expected: %s, got: %s", string(data), string(data1))
	}
}

func BenchmarkUnmarshalWithName(b *testing.B) {

	name := "countries"
	data := []byte(fmt.Sprintf(`{"%s":{"UK":{"Name":"UK","capital":"London","population":{"London":10000000}},"US":{"Name":"US","capital":"Washington DC","population":{"Washington DC":95000000}}}}`, name))

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		objs, err := UnmarshalWithName[myCountryDetails](context.Background(), name, data)
		if err != nil {
			b.Fatalf("unexpected UnmarshalWithName error: %v", err)
		}
		if len(objs) != 2 {
			b.Fatalf("unexpected unmarshal result")
		}

		data1, err := MarshalWithName(context.Background(), name, objs)
		if err != nil {
			b.Fatalf("unexpected MarshalWithName error: %v", err)
		}
		if !bytes.Equal(data, data1) {
			b.Fatal("mismatch in roundtrip")
		}
	}
}

func BenchmarkUnmarshal(b *testing.B) {

	data := []byte(`{"UK":{"Name":"UK","capital":"London","population":{"London":10000000}},"US":{"Name":"US","capital":"Washington DC","population":{"Washington DC":95000000}}}`)

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		objs, err := Unmarshal[myCountryDetails](context.Background(), data)
		if err != nil {
			b.Fatalf("unexpected Unmarshal error: %v", err)
		}
		if len(objs) != 2 {
			b.Fatalf("unexpected unmarshal result")
		}

		data1, err := Marshal(context.Background(), objs)
		if err != nil {
			b.Fatalf("unexpected Marshal error: %v", err)
		}
		if !bytes.Equal(data, data1) {
			b.Fatal("mismatch in roundtrip")
		}
	}
}

type stockHistoryMeta struct {
	Information   string `json:"1. Information"`
	Symbol        string `json:"2. Symbol"`
	LastRefreshed string `json:"3. Last Refreshed"`
	OutputSize    string `json:"4. Output Size"`
	TimeZone      string `json:"5. Time Zone"`
}

type stockHistoryElement struct {
	Date             string
	Open             string `json:"1. open"`
	High             string `json:"2. high"`
	Low              string `json:"3. low"`
	Close            string `json:"4. close"`
	AdjustedClose    string `json:"5. adjusted close"`
	Volume           string `json:"6. volume"`
	DividendAmount   string `json:"7. dividend amount"`
	SplitCoefficient string `json:"8. split coefficient"`
}

func (i *stockHistoryElement) SetName(name string) {
	i.Date = name
}

func (i *stockHistoryElement) GetName() string {
	return i.Date
}

func BenchmarkUnmarshalStructuredData(b *testing.B) {

	jsonPath := filepath.Join("example_data", "ibm_history.json")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		b.Fatalf("Failed to read JSON file: %v", err)
	}

	metaName := "Meta Data"
	dataName := "Time Series (Daily)"

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		sd, err := UnmarshalStructuredData[stockHistoryMeta, stockHistoryElement](context.Background(), metaName, dataName, data)
		if err != nil {
			b.Fatalf("unexpected UnmarshalWithName error: %v", err)
		}
		if sd.Meta == nil {
			b.Fatalf("unexpectedly received no metadata")
		}
		if len(sd.Data) == 0 {
			b.Fatalf("unexpected unmarshal result")
		}
		if sd.Meta.Symbol != "IBM" {
			b.Fatalf("mismatch in Symbol")
		}
		if sd.Data[0].Date != "1999-11-01" {
			b.Fatalf("mismatch in expected first data element: %s", sd.Data[0].Date)
		}
		if sd.Data[len(sd.Data)-1].Date != "2025-08-19" {
			b.Fatalf("mismatch in expected last data element: %s", sd.Data[len(sd.Data)-1].Date)
		}
	}
}

func BenchmarkUnmarshalStructuredData_descending(b *testing.B) {

	jsonPath := filepath.Join("example_data", "ibm_history.json")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		b.Fatalf("Failed to read JSON file: %v", err)
	}

	metaName := "Meta Data"
	dataName := "Time Series (Daily)"

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		sd, err := UnmarshalStructuredData[stockHistoryMeta](context.Background(), metaName, dataName, data, WithOrdering[stockHistoryElement](Descending))
		if err != nil {
			b.Fatalf("unexpected UnmarshalWithName error: %v", err)
		}
		if sd.Meta == nil {
			b.Fatalf("unexpectedly received no metadata")
		}
		if len(sd.Data) == 0 {
			b.Fatalf("unexpected unmarshal result")
		}
		if sd.Meta.Symbol != "IBM" {
			b.Fatalf("mismatch in Symbol")
		}
		if sd.Data[0].Date != "2025-08-19" {
			b.Fatalf("mismatch in expected last data element: %s", sd.Data[0].Date)
		}
		if sd.Data[len(sd.Data)-1].Date != "1999-11-01" {
			b.Fatalf("mismatch in expected first data element: %s", sd.Data[len(sd.Data)-1].Date)
		}
	}
}
