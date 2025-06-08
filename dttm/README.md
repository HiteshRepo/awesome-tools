# DTTM - Date Time Utilities

A Go package for flexible date/time formatting and parsing with multiple format support.

## Overview

The `dttm` package provides utilities for parsing and formatting time strings in various formats. It's particularly useful when dealing with different time representations and need to extract or convert between them.

**Note:** This package is inspired by (and largely copied from) [alcionai/corso](https://github.com/alcionai/corso/tree/main/src/pkg/dttm).

## Features

- **Multiple Time Formats**: Support for various time formats including RFC3339, human-readable formats, and custom formats
- **Flexible Parsing**: Parse time strings in any supported format automatically
- **Time Extraction**: Extract time information from strings containing other text
- **UTC Conversion**: All parsed times are automatically converted to UTC
- **Format Conversion**: Convert between different time formats easily

## Supported Formats

| Format Name | Example | Use Case |
|-------------|---------|----------|
| `Standard` | `2023-12-25T15:30:45.123456789Z` | RFC3339Nano standard |
| `DateOnly` | `2023-12-25` | Date-only representations |
| `TabularOutput` | `2023-12-25T15:30:45Z` | Clean tabular display |
| `HumanReadable` | `25-Dec-2023_15:30:45` | Human-friendly format |
| `HumanReadableDriveItem` | `25-Dec-2023_15-30-45` | Filesystem-safe format |
| `ClippedHuman` | `25-Dec-2023_15:30` | Shortened human format |
| `ClippedHumanDriveItem` | `25-Dec-2023_15-30` | Shortened filesystem-safe |
| `SafeForTesting` | `25-Dec-2023_15-30-45.000000` | Testing with microseconds |

## Installation

```bash
go get github.com/your-repo/awesome-tools/dttm
```

## Usage

### Basic Parsing

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/your-repo/awesome-tools/dttm"
)

func main() {
    // Parse a time string (automatically detects format)
    t, err := dttm.ParseTime("2023-12-25T15:30:45Z")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Parsed time:", t)
}
```

### Format Conversion

```go
package main

import (
    "fmt"
    "time"
    
    "github.com/your-repo/awesome-tools/dttm"
)

func main() {
    now := time.Now()
    
    // Format to different representations
    standard := dttm.FormatTo(now, dttm.Standard)
    human := dttm.FormatTo(now, dttm.HumanReadable)
    dateOnly := dttm.FormatTo(now, dttm.DateOnly)
    
    fmt.Printf("Standard: %s\n", standard)
    fmt.Printf("Human: %s\n", human)
    fmt.Printf("Date Only: %s\n", dateOnly)
}
```

### Extract Time from Text

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/your-repo/awesome-tools/dttm"
)

func main() {
    // Extract time from a string containing other text
    text := "Log entry from 25-Dec-2023_15:30:45 shows error"
    
    t, err := dttm.ExtractTime(text)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Extracted time: %s\n", t)
}
```

### Working with Different Formats

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/your-repo/awesome-tools/dttm"
)

func main() {
    // Parse various format examples
    examples := []string{
        "2023-12-25T15:30:45.123456789Z",
        "25-Dec-2023_15:30:45",
        "2023-12-25",
        "25-Dec-2023_15-30-45",
    }
    
    for _, example := range examples {
        t, err := dttm.ParseTime(example)
        if err != nil {
            log.Printf("Failed to parse %s: %v", example, err)
            continue
        }
        
        // Convert to different formats
        fmt.Printf("Original: %s\n", example)
        fmt.Printf("  -> Standard: %s\n", dttm.FormatTo(t, dttm.Standard))
        fmt.Printf("  -> Human: %s\n", dttm.FormatTo(t, dttm.HumanReadable))
        fmt.Printf("  -> Date Only: %s\n", dttm.FormatTo(t, dttm.DateOnly))
        fmt.Println()
    }
}
```

## API Reference

### Functions

#### `FormatTo(t time.Time, fmt TimeFormat) string`
Formats a time.Time to the specified format string. The time is converted to UTC before formatting.

**Parameters:**
- `t`: The time to format
- `fmt`: The target format (use predefined TimeFormat constants)

**Returns:** Formatted time string

#### `ParseTime(s string) (time.Time, error)`
Parses a time string using any of the supported formats. Tries each format until one succeeds.

**Parameters:**
- `s`: Time string to parse

**Returns:** 
- `time.Time`: Parsed time in UTC
- `error`: Error if parsing fails for all formats

#### `ExtractTime(s string) (time.Time, error)`
Extracts time information from a string that may contain other text. Uses regex patterns to find time substrings.

**Parameters:**
- `s`: String containing time information

**Returns:**
- `time.Time`: Extracted time in UTC  
- `error`: Error if no time pattern is found

### Types

#### `TimeFormat`
String type representing different time format patterns.

**Constants:**
- `Standard`: RFC3339Nano format
- `DateOnly`: Date-only format (YYYY-MM-DD)
- `TabularOutput`: Clean format for tables
- `HumanReadable`: Human-friendly format with colons
- `HumanReadableDriveItem`: Filesystem-safe human format
- `ClippedHuman`: Shortened human format
- `ClippedHumanDriveItem`: Shortened filesystem-safe format
- `SafeForTesting`: Testing format with microseconds

## Error Handling

The package returns descriptive errors for common failure cases:

- Empty time strings
- Unrecognized time formats
- Invalid time values

Always check for errors when parsing time strings:

```go
t, err := dttm.ParseTime(timeString)
if err != nil {
    // Handle parsing error
    log.Printf("Failed to parse time: %v", err)
    return
}
```

## Dependencies

- `github.com/pkg/errors` - Enhanced error handling
- Standard Go `time` and `regexp` packages

## License

This package is inspired by the [alcionai/corso](https://github.com/alcionai/corso) project. Please refer to the original project for licensing information.
