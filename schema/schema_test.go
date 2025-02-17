package schema

import (
	"fmt"
	"testing"
)

func TestStructToMarkdowndown(t *testing.T) {
	// Profile is a nested struct inside User.
	type Profile struct {
		Age  *int   `json:"age,omitempty" jsonschema:"title=Age,description=User's age"`
		Bio  string `json:"bio,omitempty" jsonschema:"title=Bio"`
		City string `json:"city,omitempty"`
	}

	// Tag represents an object inside the Tags array.
	type Tag struct {
		Category string              `json:"category,omitempty" jsonschema:"title=Category,description=Type of tag"`
		Tag      string              `json:"tag,omitempty" jsonschema:"title=Tag,description=Actual tag value"`
		Exps     []string            `json:"exps,omitempty" jsonschema:"title=Exps,description=Extend preferences"`
		Map      []map[string]string `json:"map,omitempty" jsonschema:"title=Map,description=Test for map"`
	}
	// User represents a sample struct with `jsonschema` tags.
	type User struct {
		Base
		ID      int                 `json:"id" jsonschema:"title=User ID,description=The unique identifier for a user"`
		Name    *string             `json:"name,omitempty" jsonschema:"title=Full Name,description=The full name of the user"`
		Email   *string             `json:"email,omitempty" jsonschema:"title=Email,description=The email address of the user"`
		Profile *Profile            `json:"profile,omitempty" jsonschema:"title=Profile,description=User profile information"`
		Tags    []*Tag              `json:"tags,omitempty" jsonschema:"title=Tags,description=List of user tags"`
		Map     map[string][]string `json:"map,omitempty" jsonschema:"title=Map,description=Test for map"`
	}
	// Example object with pointers and an array of object pointers
	name := "Alice"
	email := "alice@example.com"
	age := 30

	user := User{
		ID:    123,
		Name:  &name,
		Email: &email,
		Profile: &Profile{
			Age:  &age,
			Bio:  "Loves coding and coffee.",
			City: "Bangkok",
		},
		Tags: []*Tag{
			{Category: "Interest", Tag: "Technology", Map: []map[string]string{
				{"key1": "v1"},
				{"key2": "v2"},
			}},
			nil, // Simulating a nil pointer in the slice
			{Category: "Skill", Tag: "Golang", Exps: []string{"tag1", "tag2", "tag3"}},
		},
		Map: map[string][]string{
			"key1": {"v1", "v1.2"},
			"key2": {"v2", "v2.2"},
		},
	}

	markdown := SchemaToMarkdown(&user)
	fmt.Println(markdown)
	// fmt.Println("===================")
	// users := []User{user, user}
	// markdown = SchemaToMarkdown(&users)
	// fmt.Println(markdown)
}
