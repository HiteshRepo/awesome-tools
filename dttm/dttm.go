package dttm

// This package is inspired (rather) copied from https://github.com/alcionai/corso/tree/main/src/pkg/dttm

import (
	"regexp"
	"time"

	"github.com/pkg/errors"
)

type TimeFormat string

const (
	Standard               TimeFormat = time.RFC3339Nano
	DateOnly               TimeFormat = "2006-01-02"
	TabularOutput          TimeFormat = "2006-01-02T15:04:05Z"
	HumanReadable          TimeFormat = "02-Jan-2006_15:04:05"
	HumanReadableDriveItem TimeFormat = "02-Jan-2006_15-04-05"
	ClippedHuman           TimeFormat = "02-Jan-2006_15:04"
	ClippedHumanDriveItem  TimeFormat = "02-Jan-2006_15-04"
	SafeForTesting         TimeFormat = HumanReadableDriveItem + ".000000"
)

var (
	clippedHumanRE         = regexp.MustCompile(`.*(\d{2}-[a-zA-Z]{3}-\d{4}_\d{2}:\d{2}).*`)
	clippedHumanOneDriveRE = regexp.MustCompile(`.*(\d{2}-[a-zA-Z]{3}-\d{4}_\d{2}-\d{2}).*`)
	dateOnlyRE             = regexp.MustCompile(`.*(\d{4}-\d{2}-\d{2}).*`)
	legacyRE               = regexp.MustCompile(
		`.*(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}?([Zz]|[a-zA-Z]{2}|([\+|\-]([01]\d|2[0-3])))).*`)
	SafeForTestingRE        = regexp.MustCompile(`.*(\d{2}-[a-zA-Z]{3}-\d{4}_\d{2}-\d{2}-\d{2}.\d{6}).*`)
	HumanReadableRE         = regexp.MustCompile(`.*(\d{2}-[a-zA-Z]{3}-\d{4}_\d{2}:\d{2}:\d{2}).*`)
	HumanReadableOneDriveRE = regexp.MustCompile(`.*(\d{2}-[a-zA-Z]{3}-\d{4}_\d{2}-\d{2}-\d{2}).*`)
	standardRE              = regexp.MustCompile(
		`.*(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?([Zz]|[a-zA-Z]{2}|([\+|\-]([01]\d|2[0-3])))).*`)
	tabularOutputRE = regexp.MustCompile(`.*(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}([Zz]|[a-zA-Z]{2})).*`)
)

var (
	formats = []TimeFormat{
		Standard,
		SafeForTesting,
		HumanReadable,
		HumanReadableDriveItem,
		TabularOutput,
		ClippedHuman,
		ClippedHumanDriveItem,
		DateOnly,
	}
	regexes = []*regexp.Regexp{
		standardRE,
		SafeForTestingRE,
		HumanReadableRE,
		HumanReadableOneDriveRE,
		legacyRE,
		tabularOutputRE,
		clippedHumanRE,
		clippedHumanOneDriveRE,
		dateOnlyRE,
	}
)

func FormatTo(t time.Time, fmt TimeFormat) string {
	return t.UTC().Format(string(fmt))
}

func ParseTime(s string) (time.Time, error) {
	if len(s) == 0 {
		return time.Time{}, errors.New("empty time string passed")
	}

	var lastErr error
	var t time.Time

	for _, form := range formats {
		t, lastErr = time.Parse(string(form), s)
		if lastErr == nil {
			return t.UTC(), nil
		}
	}

	return time.Time{}, errors.Wrap(lastErr, "parsing time string")
}

func ExtractTime(s string) (time.Time, error) {
	if len(s) == 0 {
		return time.Time{}, errors.New("empty time string passed")
	}

	for _, re := range regexes {
		ss := re.FindAllStringSubmatch(s, -1)
		if len(ss) > 0 && len(ss[0]) > 1 {
			return ParseTime(ss[0][1])
		}
	}

	return time.Time{}, errors.New("no format match for provided time string")
}
