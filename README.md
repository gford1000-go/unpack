[![Go Reference](https://pkg.go.dev/badge/github.com/gford1000-go/unpack.svg)](https://pkg.go.dev/github.com/gford1000-go/unpack)

unpack | Simplified JSON decoding of a map of named objects
===========================================================

Often we come across JSON data that is presented as a map of named objects, e.g.:

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

The `unpack` package simplifies parsing of such data, including ensuring the name (key) of the data objects is captured during the parsing.

Use
===

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

func printCountries(b []byte) {
    countries, _ := Unmarshal[Country](b)
    for _, country := range country {
        fmt.Println(country.Name, country.Capital, country.Population[country.Capital])
    }
}
```

Additionally, slices of `Unpackable` can be encoded to JSON using the `Marshal` function.  Providing a name to `Marshal` will create a byte slice that can be decoded using `Unmarshal`; providing an empty string as name will create an anonymous map (note this is not currently decodable by this package).

See examples for more details.
