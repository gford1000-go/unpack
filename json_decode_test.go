package unpack

import (
	"testing"
)

func TestMapToStruct(t *testing.T) {

	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
		Zip    string `json:"zip"`
	}

	type Person struct {
		Name        string                       `json:"name"`
		Age         int                          `json:"age"`
		Email       string                       `json:"email_address"`
		Emails      map[string]string            `json:"emails"`      // map[string]any -> map[string]string
		Metadata    map[string]any               `json:"metadata"`    // stays as map[string]any
		Scores      map[string]int               `json:"scores"`      // map[string]any -> map[string]int
		Tags        []string                     `json:"tags"`        // []any -> []string
		Address     Address                      `json:"address"`     // map[string]any -> struct
		AltAddress  *Address                     `json:"alt_address"` // map[string]any -> *struct
		Preferences map[string]map[string]string `json:"preferences"` // nested maps
	}

	data := map[string]any{
		"name":          "John Doe",
		"age":           30,
		"email_address": "john@example.com",
		"emails": map[string]any{
			"work":     "john.work@example.com",
			"personal": "john.personal@example.com",
		},
		"metadata": map[string]any{
			"department": "Engineering",
			"level":      5,
		},
		"scores": map[string]any{
			"math":    95,
			"science": 87,
			"english": 92,
		},
		"tags": []any{"developer", "golang", "senior"},
		"address": map[string]any{
			"street": "123 Main St",
			"city":   "San Francisco",
			"zip":    "94105",
		},
		"alt_address": map[string]any{
			"street": "456 Oak Ave",
			"city":   "Los Angeles",
			"zip":    "90210",
		},
		"preferences": map[string]any{
			"notifications": map[string]any{
				"email": "enabled",
				"sms":   "disabled",
			},
			"theme": map[string]any{
				"color": "dark",
				"size":  "large",
			},
		},
	}

	var person Person
	err := mapToStruct(data, &person)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if person.Name != "John Doe" {
		t.Fatal("failed to convert Name")
	}
	if person.Age != 30 {
		t.Fatal("failed to covert Age")
	}
	if person.Email != "john@example.com" {
		t.Fatal("failed to convert Email")
	}
	if len(person.Emails) != 2 {
		t.Fatal("failed to convert Emails")
	}
	if v, ok := person.Emails["work"]; !ok {
		t.Fatal("Emails does not contain `work` address")
	} else if v != "john.work@example.com" {
		t.Fatal("Emails -> work is incorrect")
	}
	if v, ok := person.Emails["personal"]; !ok {
		t.Fatal("Emails does not contain `personal` address")
	} else if v != "john.personal@example.com" {
		t.Fatal("Emails -> personal is incorrect")
	}
	if len(person.Scores) != 3 {
		t.Fatal("failed to convert Scores")
	}
	var tot = 0
	for _, v := range person.Scores {
		tot += v
	}
	if tot != 274 {
		t.Fatal("failed to convert Scores - wrong total", tot)
	}
	if len(person.Tags) != 3 {
		t.Fatal("failed to convert Tags")
	}
	if person.Tags[0]+person.Tags[1]+person.Tags[2] != "developergolangsenior" {
		t.Fatal("Tags not converted")
	}
	if len(person.Preferences) != 2 {
		t.Fatal("failed to convert Preferences")
	}
	if m, ok := person.Preferences["theme"]; !ok {
		t.Fatal("Preferences missing 'theme'")
	} else {
		if len(m) != 2 {
			t.Fatal("failed to convert Preferences.Themes")
		}
	}
}
