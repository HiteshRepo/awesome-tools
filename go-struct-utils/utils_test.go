package gostructutils

import (
	"reflect"
	"testing"
)

type Person struct {
	Name     string `json:"name"`
	Age      int    `json:"age"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
}

type SimpleStruct struct {
	Name string
	Age  int
}

type StructWithTags struct {
	Name     string `json:"full_name"`
	Age      int    `json:"age"`
	Email    string `json:"email_address"`
	IsActive bool   `json:"is_active"`
}

type StructWithIgnoredFields struct {
	Name      string `json:"name"`
	Age       int    `json:"age"`
	Password  string `json:"-"`
	Internal  string `json:"-"`
	Published bool   `json:"published"`
}

type StructWithUnexported struct {
	Name     string `json:"name"`
	Age      int    `json:"age"`
	password string // unexported field
	internal int    // unexported field
}

type EmptyStruct struct{}

type NestedStruct struct {
	Person Person `json:"person"`
	Count  int    `json:"count"`
}

func mapsEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, exists := b[k]; !exists || !reflect.DeepEqual(v, bv) {
			return false
		}
	}
	return true
}

func TestStructToMapJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expected    map[string]any
		expectError bool
	}{
		{
			name:  "Simple struct",
			input: SimpleStruct{Name: "John", Age: 30},
			expected: map[string]any{
				"Name": "John",
				"Age":  float64(30),
			},
		},
		{
			name: "Struct with JSON tags",
			input: StructWithTags{
				Name:     "Jane Doe",
				Age:      25,
				Email:    "jane@example.com",
				IsActive: true,
			},
			expected: map[string]any{
				"full_name":     "Jane Doe",
				"age":           float64(25),
				"email_address": "jane@example.com",
				"is_active":     true,
			},
		},
		{
			name: "Struct with ignored fields",
			input: StructWithIgnoredFields{
				Name:      "Bob",
				Age:       35,
				Password:  "secret123",
				Internal:  "internal_data",
				Published: true,
			},
			expected: map[string]any{
				"name":      "Bob",
				"age":       float64(35),
				"published": true,
			},
		},
		{
			name: "Struct with unexported fields",
			input: StructWithUnexported{
				Name:     "Alice",
				Age:      28,
				password: "hidden",
				internal: 42,
			},
			expected: map[string]any{
				"name": "Alice",
				"age":  float64(28),
			},
		},
		{
			name:     "Empty struct",
			input:    EmptyStruct{},
			expected: map[string]any{},
		},
		{
			name: "Pointer to struct",
			input: &Person{
				Name:     "Pointer Test",
				Age:      45,
				Email:    "pointer@example.com",
				IsActive: true,
			},
			expected: map[string]any{
				"name":      "Pointer Test",
				"age":       float64(45),
				"email":     "pointer@example.com",
				"is_active": true,
			},
		},
		{
			name: "Nested struct",
			input: NestedStruct{
				Person: Person{
					Name:     "Nested Person",
					Age:      40,
					Email:    "nested@example.com",
					IsActive: false,
				},
				Count: 5,
			},
			expected: map[string]any{
				"person": map[string]any{
					"name":      "Nested Person",
					"age":       float64(40),
					"email":     "nested@example.com",
					"is_active": false,
				},
				"count": float64(5),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := StructToMapJSON(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !mapsEqual(result, tt.expected) {
				t.Errorf("Expected %+v, got %+v", tt.expected, result)
			}
		})
	}
}

func TestStructToMapReflection(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected map[string]any
	}{
		{
			name:  "Simple struct",
			input: SimpleStruct{Name: "John", Age: 30},
			expected: map[string]any{
				"Name": "John",
				"Age":  30,
			},
		},
		{
			name: "Struct with JSON tags - uses field names not tags",
			input: StructWithTags{
				Name:     "Jane Doe",
				Age:      25,
				Email:    "jane@example.com",
				IsActive: true,
			},
			expected: map[string]any{
				"full_name":     "Jane Doe",
				"age":           25,
				"email_address": "jane@example.com",
				"is_active":     true,
			},
		},
		{
			name: "Struct with ignored fields - still includes them",
			input: StructWithIgnoredFields{
				Name:      "Bob",
				Age:       35,
				Password:  "secret123",
				Internal:  "internal_data",
				Published: true,
			},
			expected: map[string]any{
				"name":      "Bob",
				"age":       35,
				"Password":  "secret123",
				"Internal":  "internal_data",
				"published": true,
			},
		},
		{
			name: "Struct with unexported fields - excludes them",
			input: StructWithUnexported{
				Name:     "Alice",
				Age:      28,
				password: "hidden",
				internal: 42,
			},
			expected: map[string]any{
				"name": "Alice",
				"age":  28,
			},
		},
		{
			name:     "Empty struct",
			input:    EmptyStruct{},
			expected: map[string]any{},
		},
		{
			name: "Pointer to struct",
			input: &Person{
				Name:     "Pointer Test",
				Age:      45,
				Email:    "pointer@example.com",
				IsActive: true,
			},
			expected: map[string]any{
				"name":      "Pointer Test",
				"age":       45,
				"email":     "pointer@example.com",
				"is_active": true,
			},
		},
		{
			name:     "Non-struct input",
			input:    "not a struct",
			expected: map[string]any{},
		},
		{
			name:     "Nil pointer",
			input:    (*Person)(nil),
			expected: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StructToMapUsingReflection(tt.input)

			if !mapsEqual(result, tt.expected) {
				t.Errorf("Expected %+v, got %+v", tt.expected, result)
			}
		})
	}
}

func TestStructToMapReflectionAdvanced(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected map[string]any
	}{
		{
			name:  "Simple struct",
			input: SimpleStruct{Name: "John", Age: 30},
			expected: map[string]any{
				"Name": "John",
				"Age":  30,
			},
		},
		{
			name: "Struct with JSON tags",
			input: StructWithTags{
				Name:     "Jane Doe",
				Age:      25,
				Email:    "jane@example.com",
				IsActive: true,
			},
			expected: map[string]any{
				"full_name":     "Jane Doe",
				"age":           25,
				"email_address": "jane@example.com",
				"is_active":     true,
			},
		},
		{
			name: "Struct with ignored fields",
			input: StructWithIgnoredFields{
				Name:      "Bob",
				Age:       35,
				Password:  "secret123",
				Internal:  "internal_data",
				Published: true,
			},
			expected: map[string]any{
				"name":      "Bob",
				"age":       35,
				"published": true,
			},
		},
		{
			name: "Struct with unexported fields",
			input: StructWithUnexported{
				Name:     "Alice",
				Age:      28,
				password: "hidden",
				internal: 42,
			},
			expected: map[string]any{
				"name": "Alice",
				"age":  28,
			},
		},
		{
			name:     "Empty struct",
			input:    EmptyStruct{},
			expected: map[string]any{},
		},
		{
			name: "Pointer to struct",
			input: &Person{
				Name:     "Pointer Test",
				Age:      45,
				Email:    "pointer@example.com",
				IsActive: true,
			},
			expected: map[string]any{
				"name":      "Pointer Test",
				"age":       45,
				"email":     "pointer@example.com",
				"is_active": true,
			},
		},
		{
			name:     "Non-struct input",
			input:    "not a struct",
			expected: map[string]any{},
		},
		{
			name:     "Nil pointer",
			input:    (*Person)(nil),
			expected: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StructToMapUsingAdvancedReflection(tt.input)

			if !mapsEqual(result, tt.expected) {
				t.Errorf("Expected %+v, got %+v", tt.expected, result)
			}
		})
	}
}
