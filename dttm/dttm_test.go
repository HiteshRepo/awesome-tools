package dttm

import (
	"testing"
	"time"
)

func TestFormatTo(t *testing.T) {
	// Create a fixed time for consistent testing
	fixedTime := time.Date(2023, 12, 25, 15, 30, 45, 123456789, time.UTC)

	tests := []struct {
		name     string
		time     time.Time
		format   TimeFormat
		expected string
	}{
		{
			name:     "Standard format",
			time:     fixedTime,
			format:   Standard,
			expected: "2023-12-25T15:30:45.123456789Z",
		},
		{
			name:     "DateOnly format",
			time:     fixedTime,
			format:   DateOnly,
			expected: "2023-12-25",
		},
		{
			name:     "TabularOutput format",
			time:     fixedTime,
			format:   TabularOutput,
			expected: "2023-12-25T15:30:45Z",
		},
		{
			name:     "HumanReadable format",
			time:     fixedTime,
			format:   HumanReadable,
			expected: "25-Dec-2023_15:30:45",
		},
		{
			name:     "HumanReadableDriveItem format",
			time:     fixedTime,
			format:   HumanReadableDriveItem,
			expected: "25-Dec-2023_15-30-45",
		},
		{
			name:     "ClippedHuman format",
			time:     fixedTime,
			format:   ClippedHuman,
			expected: "25-Dec-2023_15:30",
		},
		{
			name:     "ClippedHumanDriveItem format",
			time:     fixedTime,
			format:   ClippedHumanDriveItem,
			expected: "25-Dec-2023_15-30",
		},
		{
			name:     "SafeForTesting format",
			time:     fixedTime,
			format:   SafeForTesting,
			expected: "25-Dec-2023_15-30-45.123456",
		},
		{
			name:     "Zero time with Standard format",
			time:     time.Time{},
			format:   Standard,
			expected: "0001-01-01T00:00:00Z",
		},
		{
			name:     "Time with different timezone gets converted to UTC",
			time:     time.Date(2023, 12, 25, 15, 30, 45, 0, time.FixedZone("EST", -5*3600)),
			format:   Standard,
			expected: "2023-12-25T20:30:45Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTo(tt.time, tt.format)
			if result != tt.expected {
				t.Errorf("FormatTo() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		expected    time.Time
	}{
		{
			name:        "Standard RFC3339Nano format",
			input:       "2023-12-25T15:30:45.123456789Z",
			expectError: false,
			expected:    time.Date(2023, 12, 25, 15, 30, 45, 123456789, time.UTC),
		},
		{
			name:        "Standard RFC3339 format",
			input:       "2023-12-25T15:30:45Z",
			expectError: false,
			expected:    time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC),
		},
		{
			name:        "DateOnly format",
			input:       "2023-12-25",
			expectError: false,
			expected:    time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC),
		},
		{
			name:        "TabularOutput format",
			input:       "2023-12-25T15:30:45Z",
			expectError: false,
			expected:    time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC),
		},
		{
			name:        "HumanReadable format",
			input:       "25-Dec-2023_15:30:45",
			expectError: false,
			expected:    time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC),
		},
		{
			name:        "HumanReadableDriveItem format",
			input:       "25-Dec-2023_15-30-45",
			expectError: false,
			expected:    time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC),
		},
		{
			name:        "ClippedHuman format",
			input:       "25-Dec-2023_15:30",
			expectError: false,
			expected:    time.Date(2023, 12, 25, 15, 30, 0, 0, time.UTC),
		},
		{
			name:        "ClippedHumanDriveItem format",
			input:       "25-Dec-2023_15-30",
			expectError: false,
			expected:    time.Date(2023, 12, 25, 15, 30, 0, 0, time.UTC),
		},
		{
			name:        "SafeForTesting format",
			input:       "25-Dec-2023_15-30-45.000000",
			expectError: false,
			expected:    time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC),
		},
		{
			name:        "SafeForTesting format with microseconds",
			input:       "25-Dec-2023_15-30-45.123456",
			expectError: false,
			expected:    time.Date(2023, 12, 25, 15, 30, 45, 123456000, time.UTC),
		},
		{
			name:        "RFC3339 with timezone offset",
			input:       "2023-12-25T15:30:45+05:30",
			expectError: false,
			expected:    time.Date(2023, 12, 25, 10, 0, 45, 0, time.UTC),
		},
		{
			name:        "RFC3339 with negative timezone offset",
			input:       "2023-12-25T15:30:45-05:00",
			expectError: false,
			expected:    time.Date(2023, 12, 25, 20, 30, 45, 0, time.UTC),
		},
		{
			name:        "Empty string",
			input:       "",
			expectError: true,
		},
		{
			name:        "Invalid format",
			input:       "not-a-time",
			expectError: true,
		},
		{
			name:        "Invalid date",
			input:       "2023-13-45T25:70:90Z",
			expectError: true,
		},
		{
			name:        "Partial time string",
			input:       "2023-12",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTime(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("ParseTime() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseTime() unexpected error: %v", err)
				return
			}

			if !result.Equal(tt.expected) {
				t.Errorf("ParseTime() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractTime(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		expected    time.Time
	}{
		{
			name:        "Extract Standard format from text",
			input:       "Log entry at 2023-12-25T15:30:45.123456789Z shows error",
			expectError: false,
			expected:    time.Date(2023, 12, 25, 15, 30, 45, 123456789, time.UTC),
		},
		{
			name:        "Extract HumanReadable format from text",
			input:       "File created on 25-Dec-2023_15:30:45 successfully",
			expectError: false,
			expected:    time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC),
		},
		{
			name:        "Extract HumanReadableDriveItem format from text",
			input:       "Backup from 25-Dec-2023_15-30-45 completed",
			expectError: false,
			expected:    time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC),
		},
		{
			name:        "Extract DateOnly format from text",
			input:       "Report for date 2023-12-25 is ready",
			expectError: false,
			expected:    time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC),
		},
		{
			name:        "Extract ClippedHuman format from text",
			input:       "Meeting scheduled for 25-Dec-2023_15:30 today",
			expectError: false,
			expected:    time.Date(2023, 12, 25, 15, 30, 0, 0, time.UTC),
		},
		{
			name:        "Extract ClippedHumanDriveItem format from text",
			input:       "Version 25-Dec-2023_15-30 released",
			expectError: false,
			expected:    time.Date(2023, 12, 25, 15, 30, 0, 0, time.UTC),
		},
		{
			name:        "Extract SafeForTesting format from text",
			input:       "Test run 25-Dec-2023_15-30-45.123456 passed",
			expectError: false,
			expected:    time.Date(2023, 12, 25, 15, 30, 45, 123456000, time.UTC),
		},
		{
			name:        "Extract TabularOutput format from text",
			input:       "Event occurred at 2023-12-25T15:30:45Z in the system",
			expectError: false,
			expected:    time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC),
		},
		{
			name:        "Extract time with timezone from text",
			input:       "Timestamp: 2023-12-25T15:30:45Z for reference",
			expectError: false,
			expected:    time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC),
		},
		{
			name:        "Multiple time formats in text - should extract first match",
			input:       "Start: 2023-12-25T15:30:45Z End: 25-Dec-2023_16:30:45",
			expectError: false,
			expected:    time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC),
		},
		{
			name:        "Time at beginning of string",
			input:       "2023-12-25T15:30:45Z: System started",
			expectError: false,
			expected:    time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC),
		},
		{
			name:        "Time at end of string",
			input:       "System shutdown at 2023-12-25T15:30:45Z",
			expectError: false,
			expected:    time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC),
		},
		{
			name:        "Empty string",
			input:       "",
			expectError: true,
		},
		{
			name:        "No time pattern in string",
			input:       "This is just a regular text without any time information",
			expectError: true,
		},
		{
			name:        "Invalid time pattern",
			input:       "Something like 2023-25-45T99:99:99Z but invalid",
			expectError: true,
		},
		{
			name:        "Partial time pattern",
			input:       "Incomplete time 2023-12 here",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractTime(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("ExtractTime() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ExtractTime() unexpected error: %v", err)
				return
			}

			if !result.Equal(tt.expected) {
				t.Errorf("ExtractTime() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTimeFormatConstants(t *testing.T) {
	tests := []struct {
		name     string
		format   TimeFormat
		expected string
	}{
		{
			name:     "Standard format constant",
			format:   Standard,
			expected: time.RFC3339Nano,
		},
		{
			name:     "DateOnly format constant",
			format:   DateOnly,
			expected: "2006-01-02",
		},
		{
			name:     "TabularOutput format constant",
			format:   TabularOutput,
			expected: "2006-01-02T15:04:05Z",
		},
		{
			name:     "HumanReadable format constant",
			format:   HumanReadable,
			expected: "02-Jan-2006_15:04:05",
		},
		{
			name:     "HumanReadableDriveItem format constant",
			format:   HumanReadableDriveItem,
			expected: "02-Jan-2006_15-04-05",
		},
		{
			name:     "ClippedHuman format constant",
			format:   ClippedHuman,
			expected: "02-Jan-2006_15:04",
		},
		{
			name:     "ClippedHumanDriveItem format constant",
			format:   ClippedHumanDriveItem,
			expected: "02-Jan-2006_15-04",
		},
		{
			name:     "SafeForTesting format constant",
			format:   SafeForTesting,
			expected: "02-Jan-2006_15-04-05.000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.format) != tt.expected {
				t.Errorf("TimeFormat constant %s = %v, want %v", tt.name, string(tt.format), tt.expected)
			}
		})
	}
}

func TestRoundTripConversion(t *testing.T) {
	// Test that we can format a time and then parse it back to get the same result
	originalTime := time.Date(2023, 12, 25, 15, 30, 45, 123456000, time.UTC)

	tests := []struct {
		name   string
		format TimeFormat
	}{
		{"Standard", Standard},
		{"DateOnly", DateOnly},
		{"TabularOutput", TabularOutput},
		{"HumanReadable", HumanReadable},
		{"HumanReadableDriveItem", HumanReadableDriveItem},
		{"ClippedHuman", ClippedHuman},
		{"ClippedHumanDriveItem", ClippedHumanDriveItem},
		{"SafeForTesting", SafeForTesting},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Format the time
			formatted := FormatTo(originalTime, tt.format)

			// Parse it back
			parsed, err := ParseTime(formatted)
			if err != nil {
				t.Errorf("Failed to parse formatted time: %v", err)
				return
			}

			// For DateOnly format, we expect only the date part to match
			if tt.format == DateOnly {
				expectedDate := time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)
				if !parsed.Equal(expectedDate) {
					t.Errorf("Round trip failed for %s: got %v, want %v", tt.name, parsed, expectedDate)
				}
				return
			}

			// For clipped formats, we expect precision loss
			if tt.format == ClippedHuman || tt.format == ClippedHumanDriveItem {
				expectedClipped := time.Date(2023, 12, 25, 15, 30, 0, 0, time.UTC)
				if !parsed.Equal(expectedClipped) {
					t.Errorf("Round trip failed for %s: got %v, want %v", tt.name, parsed, expectedClipped)
				}
				return
			}

			// For SafeForTesting, we expect microsecond precision
			if tt.format == SafeForTesting {
				expectedSafe := time.Date(2023, 12, 25, 15, 30, 45, 123456000, time.UTC)
				if !parsed.Equal(expectedSafe) {
					t.Errorf("Round trip failed for %s: got %v, want %v", tt.name, parsed, expectedSafe)
				}
				return
			}

			// For TabularOutput, we expect second precision
			if tt.format == TabularOutput {
				expectedTabular := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)
				if !parsed.Equal(expectedTabular) {
					t.Errorf("Round trip failed for %s: got %v, want %v", tt.name, parsed, expectedTabular)
				}
				return
			}

			// For HumanReadable formats, we expect second precision (no nanoseconds)
			if tt.format == HumanReadable || tt.format == HumanReadableDriveItem {
				expectedSecond := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)
				if !parsed.Equal(expectedSecond) {
					t.Errorf("Round trip failed for %s: got %v, want %v", tt.name, parsed, expectedSecond)
				}
				return
			}

			// For other formats, we expect the full precision to be maintained
			if !parsed.Equal(originalTime) {
				t.Errorf("Round trip failed for %s: got %v, want %v", tt.name, parsed, originalTime)
			}
		})
	}
}

func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		testFunc    func() error
		expectError bool
	}{
		{
			name: "ParseTime with whitespace",
			testFunc: func() error {
				_, err := ParseTime("  2023-12-25T15:30:45Z  ")
				return err
			},
			expectError: true, // Should fail because of leading/trailing whitespace
		},
		{
			name: "ExtractTime with multiple valid patterns",
			testFunc: func() error {
				// Should extract the first valid pattern found
				result, err := ExtractTime("First: 2023-12-25 Second: 25-Dec-2023_15:30:45")
				if err != nil {
					return err
				}
				expected := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)
				if !result.Equal(expected) {
					t.Errorf("Expected first pattern to be extracted")
				}
				return nil
			},
			expectError: false,
		},
		{
			name: "FormatTo with far future date",
			testFunc: func() error {
				farFuture := time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC)
				result := FormatTo(farFuture, Standard)
				if result == "" {
					t.Errorf("FormatTo should handle far future dates")
				}
				return nil
			},
			expectError: false,
		},
		{
			name: "FormatTo with far past date",
			testFunc: func() error {
				farPast := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
				result := FormatTo(farPast, Standard)
				if result == "" {
					t.Errorf("FormatTo should handle far past dates")
				}
				return nil
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.testFunc()
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
