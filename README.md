unpack
======

unpack provides a convenient mechanism to read a JSON config file comprising an single object with attributes representing
unique values of the same structure:

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

## Use

Two `struct` types are required:

- A receiving `struct` that includes the attributes and associated `json` tags necessary to parse the data successfully - i.e. as required to use `json.Unmarshal`.  In addition, this struct must have a pointer type implementation of `SetName` so that the struct implements the interface `Unpackable`.
- A factory `struct` that can manufacture pointer instances of the receiving `struct` 

```go
type Country struct {
	Name string
	Capital string `json:"capital"`
	Population map[string]int `json:"population"`
}

func (c *Country) SetName(name string) {
	c.Name = name
}

type CountryFact struct{}

func (f CountryFact) New() Unpackable {
	return new(Country)
}
```

## How?

The command line is all you need.

```
go get github.com/gford1000-go/unpack
```
