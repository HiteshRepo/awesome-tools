# Go Struct Utils

A Go package providing utilities for converting Go structs to maps using different approaches.

## Overview

The `gostructutils` package offers multiple methods to convert Go structs to `map[string]any`, each with different characteristics and use cases. This is particularly useful for serialization, API responses, dynamic data processing, and working with JSON-like data structures.

## Features

- **Multiple Conversion Methods**: Three different approaches for struct-to-map conversion
- **JSON Tag Support**: Respects JSON struct tags for field naming and omission
- **Reflection-Based**: Uses Go's reflection capabilities for runtime struct analysis
- **Flexible Field Handling**: Different strategies for handling unexported fields and JSON tags
- **Comprehensive Testing**: Well-tested with various struct types and edge cases

## Installation

```bash
go get github.com/your-repo/awesome-tools/go-struct-utils
```

## Available Functions

### 1. `StructToMapJSON(s any) (map[string]any, error)`

Converts a struct to a map using JSON marshaling/unmarshaling. This method:
- **Respects JSON tags** completely (including omitempty, field renaming, and `-` for exclusion)
- **Handles nested structs** properly
- **Returns numeric values as float64** (JSON behavior)
- **May return errors** if marshaling fails

**Best for**: When you need full JSON compatibility and proper handling of JSON tags.

### 2. `StructToMapUsingReflection(s any) map[string]any`

Converts a struct to a map using basic reflection. This method:
- **Uses JSON tag names** when available, falls back to field names
- **Ignores unexported fields** automatically
- **Does NOT respect `-` JSON tags** (includes all exported fields)
- **Preserves original Go types** (int stays int, not float64)
- **Never returns errors**

**Best for**: When you want simple reflection-based conversion with basic JSON tag support.

### 3. `StructToMapUsingAdvancedReflection(s any) map[string]any`

Converts a struct to a map using advanced reflection with full JSON tag parsing. This method:
- **Fully respects JSON tags** including `-` for exclusion
- **Parses JSON tag options** properly (handles commas in tags)
- **Ignores unexported fields** automatically
- **Preserves original Go types** (int stays int, not float64)
- **Never returns errors**

**Best for**: When you want reflection-based conversion with complete JSON tag compliance.

## Usage Examples

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    
    gostructutils "github.com/your-repo/awesome-tools/go-struct-utils"
)

type Person struct {
    Name     string `json:"name"`
    Age      int    `json:"age"`
    Email    string `json:"email"`
    IsActive bool   `json:"is_active"`
}

func main() {
    person := Person{
        Name:     "John Doe",
        Age:      30,
        Email:    "john@example.com",
        IsActive: true,
    }
    
    // Method 1: JSON-based conversion
    result1, err := gostructutils.StructToMapJSON(person)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("JSON method: %+v\n", result1)
    // Output: map[age:30 email:john@example.com is_active:true name:John Doe]
    
    // Method 2: Basic reflection
    result2 := gostructutils.StructToMapUsingReflection(person)
    fmt.Printf("Basic reflection: %+v\n", result2)
    // Output: map[age:30 email:john@example.com is_active:true name:John Doe]
    
    // Method 3: Advanced reflection
    result3 := gostructutils.StructToMapUsingAdvancedReflection(person)
    fmt.Printf("Advanced reflection: %+v\n", result3)
    // Output: map[age:30 email:john@example.com is_active:true name:John Doe]
}
```

### Handling JSON Tags and Field Exclusion

```go
package main

import (
    "fmt"
    
    gostructutils "github.com/your-repo/awesome-tools/go-struct-utils"
)

type User struct {
    Name      string `json:"full_name"`
    Age       int    `json:"age"`
    Password  string `json:"-"`           // Excluded from JSON
    Internal  string `json:"-"`           // Excluded from JSON
    Published bool   `json:"published"`
}

func main() {
    user := User{
        Name:      "Jane Smith",
        Age:       25,
        Password:  "secret123",
        Internal:  "internal_data",
        Published: true,
    }
    
    // JSON method - respects all JSON tags
    result1, _ := gostructutils.StructToMapJSON(user)
    fmt.Printf("JSON method: %+v\n", result1)
    // Output: map[age:25 full_name:Jane Smith published:true]
    
    // Basic reflection - ignores "-" tags
    result2 := gostructutils.StructToMapUsingReflection(user)
    fmt.Printf("Basic reflection: %+v\n", result2)
    // Output: map[Internal:internal_data Password:secret123 age:25 full_name:Jane Smith published:true]
    
    // Advanced reflection - respects "-" tags
    result3 := gostructutils.StructToMapUsingAdvancedReflection(user)
    fmt.Printf("Advanced reflection: %+v\n", result3)
    // Output: map[age:25 full_name:Jane Smith published:true]
}
```

### Working with Nested Structs

```go
package main

import (
    "fmt"
    
    gostructutils "github.com/your-repo/awesome-tools/go-struct-utils"
)

type Address struct {
    Street string `json:"street"`
    City   string `json:"city"`
}

type Person struct {
    Name    string  `json:"name"`
    Age     int     `json:"age"`
    Address Address `json:"address"`
}

func main() {
    person := Person{
        Name: "Alice Johnson",
        Age:  28,
        Address: Address{
            Street: "123 Main St",
            City:   "New York",
        },
    }
    
    // JSON method handles nested structs properly
    result, _ := gostructutils.StructToMapJSON(person)
    fmt.Printf("Nested struct: %+v\n", result)
    // Output: map[address:map[city:New York street:123 Main St] age:28 name:Alice Johnson]
}
```

### Working with Pointers

```go
package main

import (
    "fmt"
    
    gostructutils "github.com/your-repo/awesome-tools/go-struct-utils"
)

type Person struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

func main() {
    person := &Person{
        Name: "Bob Wilson",
        Age:  35,
    }
    
    // All methods handle pointers automatically
    result, _ := gostructutils.StructToMapJSON(person)
    fmt.Printf("Pointer to struct: %+v\n", result)
    // Output: map[age:35 name:Bob Wilson]
}
```

## Method Comparison

| Feature | JSON Method | Basic Reflection | Advanced Reflection |
|---------|-------------|------------------|-------------------|
| JSON tag names | ✅ Full support | ✅ Basic support | ✅ Full support |
| JSON tag exclusion (`-`) | ✅ Respected | ❌ Ignored | ✅ Respected |
| Nested structs | ✅ Proper handling | ⚠️ As struct values | ⚠️ As struct values |
| Type preservation | ❌ Numbers → float64 | ✅ Original types | ✅ Original types |
| Error handling | ✅ Can return errors | ❌ Never errors | ❌ Never errors |
| Performance | ⚠️ Slower (marshal/unmarshal) | ✅ Fast | ✅ Fast |
| Unexported fields | ✅ Excluded | ✅ Excluded | ✅ Excluded |

## When to Use Each Method

### Use `StructToMapJSON` when:
- You need full JSON compatibility
- Working with nested structs that should be flattened to maps
- You want complete JSON tag compliance
- Performance is not critical
- You can handle potential marshaling errors

### Use `StructToMapUsingReflection` when:
- You want simple, fast conversion
- You need to preserve original Go types
- You don't need strict JSON tag exclusion
- You want basic JSON tag name mapping

### Use `StructToMapUsingAdvancedReflection` when:
- You want the best of both worlds: speed + JSON tag compliance
- You need to preserve original Go types
- You want full JSON tag support without JSON marshaling overhead
- You need reliable, error-free conversion

## Error Handling

Only `StructToMapJSON` can return errors. Always check for errors when using this method:

```go
result, err := gostructutils.StructToMapJSON(myStruct)
if err != nil {
    log.Printf("Failed to convert struct: %v", err)
    return
}
```

The reflection-based methods never return errors and will return an empty map for invalid inputs.

## Edge Cases

### Non-struct Input
```go
// All methods handle non-struct input gracefully
result := gostructutils.StructToMapUsingReflection("not a struct")
// Returns: map[string]any{}
```

### Nil Pointers
```go
var person *Person = nil
result := gostructutils.StructToMapUsingReflection(person)
// Returns: map[string]any{}
```

### Empty Structs
```go
type Empty struct{}
result := gostructutils.StructToMapUsingReflection(Empty{})
// Returns: map[string]any{}
```

## Testing

The package includes comprehensive tests covering:
- Various struct types and configurations
- JSON tag handling
- Pointer and nil handling
- Edge cases and error conditions
- All three conversion methods

Run tests:
```bash
go test ./go-struct-utils
```

## Dependencies

- Standard Go `encoding/json` package
- Standard Go `reflect` package

No external dependencies required.

## Performance Considerations

- **JSON method**: Slower due to marshal/unmarshal overhead, but handles complex nested structures
- **Reflection methods**: Faster direct field access, suitable for high-performance scenarios
- **Memory usage**: All methods create new maps; consider reusing maps for high-frequency operations

## License

This package is part of the awesome-tools collection. See the main repository for licensing information.
