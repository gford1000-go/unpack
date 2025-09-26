[![Go Reference](https://pkg.go.dev/badge/github.com/gford1000-go/unpack.svg)](https://pkg.go.dev/github.com/gford1000-go/unpack)

unpack | Simplified JSON decoding of a map of named objects
===========================================================

Often we come across JSON data that is presented either as a named map of named objects, e.g.:

```json
{
  "countries": {
    "United Kingdom": { 
      "capital": "London",
      "population": {
        "2023": 66000000
      }
    },
    "United States": {
      "capital": "Washington",
      "population": {
        "2023": 314000000
      }
    }
  }
}
```

or an "anonymous" map, containing just the named objects, e.g. :

```json
{
  "United Kingdom": { 
    "capital": "London",
    "population": {
      "2023": 66000000
    }
  },
  "United States": {
    "capital": "Washington",
    "population": {
      "2023": 314000000
    }
  }
}
```

The `unpack` package simplifies parsing of such data, including ensuring the name (key) of the data objects is captured during the parsing.

Usage
=====

Define a receiving `struct` type that includes the attributes and associated `json` tags necessary to parse the data successfully - i.e. as required to use `json.Unmarshal`.  

This type must implement the `Unpackable` interface, which requires two functions, `GetName` and `SetName`.

```go
type Country struct {
    Name string
    Capital string `json:"capital"`
    Population map[string]int `json:"population"`
}

func (c *Country) SetName(name string) {
    c.Name = name
}

func (c *Country) GetName() string {
    return c.Name
}

func main(b []byte) {
  ctx := context.Background()

  b := []byte(`{"UK":{"capital":"London","population":{"London":12000000}},"US":{"capital":"Washington DC","population":{"Washington DC":9500000}}}`)

  countries, _ := unpack.Unmarshal[myCountryDetails](ctx, b)
  for _, country := range countries {
    fmt.Println(country.Name, country.Capital, country.Population[country.Capital])
  }
}
```

The `Marshal` and `MarshalWithName` functions will JSON encode the provided objects, as
an anonymous or named map of maps respectively, providing full round trip.

See examples for more details.

Monitoring
==========

Open-Telemetry spans will be created for marshaling and unmarshaling if the context has an active trace.

Span names are `unpack.marshal` and `unpack.unmarshal` repectively.
