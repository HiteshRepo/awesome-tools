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

// Additional test structs for better coverage
type StructWithComplexTags struct {
	Name        string  `json:"name,omitempty"`
	Age         int     `json:"age,omitempty"`
	Score       float64 `json:"score,omitempty"`
	IsActive    bool    `json:"is_active,omitempty"`
	Description string  `json:"description,omitempty"`
}

type StructWithEmptyJSONTag struct {
	Name string `json:""`
	Age  int    `json:""`
}

type StructWithOnlyCommaTag struct {
	Name string `json:",omitempty"`
	Age  int    `json:","`
}

type StructWithVariousTypes struct {
	StringField    string         `json:"string_field"`
	IntField       int            `json:"int_field"`
	FloatField     float64        `json:"float_field"`
	BoolField      bool           `json:"bool_field"`
	SliceField     []string       `json:"slice_field"`
	MapField       map[string]int `json:"map_field"`
	InterfaceField interface{}    `json:"interface_field"`
	PointerField   *string        `json:"pointer_field"`
}

type CircularA struct {
	Name string     `json:"name"`
	B    *CircularB `json:"b"`
}

type CircularB struct {
	Name string     `json:"name"`
	A    *CircularA `json:"a"`
}

type UnmarshalableStruct struct {
	Channel chan int `json:"channel"`
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
		{
			name: "Struct with various types",
			input: StructWithVariousTypes{
				StringField:    "test",
				IntField:       42,
				FloatField:     3.14,
				BoolField:      true,
				SliceField:     []string{"a", "b", "c"},
				MapField:       map[string]int{"key1": 1, "key2": 2},
				InterfaceField: "interface_value",
				PointerField:   stringPtr("pointer_value"),
			},
			expected: map[string]any{
				"string_field":    "test",
				"int_field":       float64(42),
				"float_field":     3.14,
				"bool_field":      true,
				"slice_field":     []interface{}{"a", "b", "c"},
				"map_field":       map[string]interface{}{"key1": float64(1), "key2": float64(2)},
				"interface_field": "interface_value",
				"pointer_field":   "pointer_value",
			},
		},
		{
			name: "Struct with nil pointer field",
			input: StructWithVariousTypes{
				StringField:  "test",
				PointerField: nil,
			},
			expected: map[string]any{
				"string_field":    "test",
				"int_field":       float64(0),
				"float_field":     float64(0),
				"bool_field":      false,
				"slice_field":     nil,
				"map_field":       nil,
				"interface_field": nil,
				"pointer_field":   nil,
			},
		},
		{
			name:        "Unmarshalable struct with channel",
			input:       UnmarshalableStruct{Channel: make(chan int)},
			expectError: true,
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: nil,
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
		{
			name: "Struct with empty JSON tag",
			input: StructWithEmptyJSONTag{
				Name: "Test",
				Age:  25,
			},
			expected: map[string]any{
				"Name": "Test",
				"Age":  25,
			},
		},
		{
			name: "Struct with various types",
			input: StructWithVariousTypes{
				StringField:    "test",
				IntField:       42,
				FloatField:     3.14,
				BoolField:      true,
				SliceField:     []string{"a", "b", "c"},
				MapField:       map[string]int{"key1": 1, "key2": 2},
				InterfaceField: "interface_value",
				PointerField:   stringPtr("pointer_value"),
			},
			expected: map[string]any{
				"string_field":    "test",
				"int_field":       42,
				"float_field":     3.14,
				"bool_field":      true,
				"slice_field":     []string{"a", "b", "c"},
				"map_field":       map[string]int{"key1": 1, "key2": 2},
				"interface_field": "interface_value",
				"pointer_field":   stringPtr("pointer_value"),
			},
		},
		{
			name:     "Integer input",
			input:    42,
			expected: map[string]any{},
		},
		{
			name:     "Slice input",
			input:    []string{"a", "b", "c"},
			expected: map[string]any{},
		},
		{
			name:     "Map input",
			input:    map[string]int{"key": 1},
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
		{
			name: "Struct with empty JSON tag",
			input: StructWithEmptyJSONTag{
				Name: "Test",
				Age:  25,
			},
			expected: map[string]any{
				"Name": "Test",
				"Age":  25,
			},
		},
		{
			name: "Struct with comma-only JSON tag",
			input: StructWithOnlyCommaTag{
				Name: "Test",
				Age:  25,
			},
			expected: map[string]any{
				"Name": "Test",
				"Age":  25,
			},
		},
		{
			name: "Struct with complex JSON tags",
			input: StructWithComplexTags{
				Name:        "Test",
				Age:         25,
				Score:       95.5,
				IsActive:    true,
				Description: "A test description",
			},
			expected: map[string]any{
				"name":        "Test",
				"age":         25,
				"score":       95.5,
				"is_active":   true,
				"description": "A test description",
			},
		},
		{
			name: "Struct with various types",
			input: StructWithVariousTypes{
				StringField:    "test",
				IntField:       42,
				FloatField:     3.14,
				BoolField:      true,
				SliceField:     []string{"a", "b", "c"},
				MapField:       map[string]int{"key1": 1, "key2": 2},
				InterfaceField: "interface_value",
				PointerField:   stringPtr("pointer_value"),
			},
			expected: map[string]any{
				"string_field":    "test",
				"int_field":       42,
				"float_field":     3.14,
				"bool_field":      true,
				"slice_field":     []string{"a", "b", "c"},
				"map_field":       map[string]int{"key1": 1, "key2": 2},
				"interface_field": "interface_value",
				"pointer_field":   stringPtr("pointer_value"),
			},
		},
		{
			name:     "Integer input",
			input:    42,
			expected: map[string]any{},
		},
		{
			name:     "Slice input",
			input:    []string{"a", "b", "c"},
			expected: map[string]any{},
		},
		{
			name:     "Map input",
			input:    map[string]int{"key": 1},
			expected: map[string]any{},
		},
		{
			name:     "Interface input",
			input:    interface{}("test"),
			expected: map[string]any{},
		},
		{
			name:     "Function input",
			input:    func() {},
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

// Test edge cases for JSON tag parsing in basic reflection
func TestStructToMapReflectionJSONTagParsing(t *testing.T) {
	type TestStruct struct {
		Field1 string `json:"field1,omitempty"`
		Field2 string `json:"field2,omitempty,string"`
		Field3 string `json:"field3"`
		Field4 string `json:",omitempty"`
		Field5 string `json:""`
		Field6 string // no tag
	}

	input := TestStruct{
		Field1: "value1",
		Field2: "value2",
		Field3: "value3",
		Field4: "value4",
		Field5: "value5",
		Field6: "value6",
	}

	result := StructToMapUsingReflection(input)
	expected := map[string]any{
		"field1": "value1",
		"field2": "value2",
		"field3": "value3",
		"":       "value4", // comma-only tag results in empty string key
		"Field5": "value5", // empty tag should use field name
		"Field6": "value6", // no tag should use field name
	}

	if !mapsEqual(result, expected) {
		t.Errorf("Expected %+v, got %+v", expected, result)
	}
}

// Test edge cases for JSON tag parsing in advanced reflection
func TestStructToMapAdvancedReflectionJSONTagParsing(t *testing.T) {
	type TestStruct struct {
		Field1 string `json:"field1,omitempty"`
		Field2 string `json:"field2,omitempty,string"`
		Field3 string `json:"field3"`
		Field4 string `json:",omitempty"`
		Field5 string `json:""`
		Field6 string // no tag
		Field7 string `json:"field7,omitempty,required"`
	}

	input := TestStruct{
		Field1: "value1",
		Field2: "value2",
		Field3: "value3",
		Field4: "value4",
		Field5: "value5",
		Field6: "value6",
		Field7: "value7",
	}

	result := StructToMapUsingAdvancedReflection(input)
	expected := map[string]any{
		"field1": "value1",
		"field2": "value2",
		"field3": "value3",
		"Field4": "value4", // comma-only tag should use field name
		"Field5": "value5", // empty tag should use field name
		"Field6": "value6", // no tag should use field name
		"field7": "value7",
	}

	if !mapsEqual(result, expected) {
		t.Errorf("Expected %+v, got %+v", expected, result)
	}
}

// Test nil pointer handling
func TestNilPointerHandling(t *testing.T) {
	var nilPerson *Person = nil
	var nilInterface interface{} = nil

	// Test JSON method
	result1, err := StructToMapJSON(nilPerson)
	if err != nil {
		t.Errorf("StructToMapJSON with nil pointer should not error, got: %v", err)
	}
	if result1 != nil {
		t.Errorf("StructToMapJSON with nil pointer should return nil, got: %+v", result1)
	}

	result2, err := StructToMapJSON(nilInterface)
	if err != nil {
		t.Errorf("StructToMapJSON with nil interface should not error, got: %v", err)
	}
	if result2 != nil {
		t.Errorf("StructToMapJSON with nil interface should return nil, got: %+v", result2)
	}

	// Test reflection methods
	result3 := StructToMapUsingReflection(nilPerson)
	expected := map[string]any{}
	if !mapsEqual(result3, expected) {
		t.Errorf("StructToMapUsingReflection with nil pointer should return empty map, got: %+v", result3)
	}

	result4 := StructToMapUsingAdvancedReflection(nilPerson)
	if !mapsEqual(result4, expected) {
		t.Errorf("StructToMapUsingAdvancedReflection with nil pointer should return empty map, got: %+v", result4)
	}
}

// Benchmark tests
func BenchmarkStructToMapJSON(b *testing.B) {
	person := Person{
		Name:     "John Doe",
		Age:      30,
		Email:    "john@example.com",
		IsActive: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = StructToMapJSON(person)
	}
}

func BenchmarkStructToMapUsingReflection(b *testing.B) {
	person := Person{
		Name:     "John Doe",
		Age:      30,
		Email:    "john@example.com",
		IsActive: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = StructToMapUsingReflection(person)
	}
}

func BenchmarkStructToMapUsingAdvancedReflection(b *testing.B) {
	person := Person{
		Name:     "John Doe",
		Age:      30,
		Email:    "john@example.com",
		IsActive: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = StructToMapUsingAdvancedReflection(person)
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
